package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

//go:embed scripts/*
var scripts embed.FS

// Browser-specific AppleScript snippets to get the URL of the active tab in the browser
var browserUrlScripts = map[string]string{
	"Arc":           "tell application \"Arc\" to set currentURL to URL of active tab of front window",
	"Google Chrome": "tell application \"Google Chrome\" to set currentURL to URL of active tab of front window",
	"Safari":        "tell application \"Safari\" to set currentURL to URL of current tab of window 1",
}

// Run any AppleScript and return the output
func runAppleScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("AppleScript error: %v, stderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

type SystemStatus struct {
	IsScreenSaverRunning bool `json:"isScreenSaverRunning"`
	IsLocked             bool `json:"isLocked"`
	IsAsleep             bool `json:"isAsleep"`
}

// Get the name of the frontmost application and the title of the frontmost window
func getAppAndWindow() (string, string, error) {
	// Check if display is asleep, screen saver is running, or the screen is locked
	data, _ := scripts.ReadFile("scripts/sleep.scpt")
	sleepStr, err := runAppleScript(string(data))
	if err != nil {
		return "", "", err
	}
	var systemStatus SystemStatus
	if err := json.Unmarshal([]byte(sleepStr), &systemStatus); err != nil {
		return "", "", err
	}
	if systemStatus.IsScreenSaverRunning || systemStatus.IsLocked || systemStatus.IsAsleep {
		return "", "", nil
	}

	// Get currently active application name
	appName, err := runAppleScript(`tell application "System Events" to get name of first process whose frontmost is true`)
	if err != nil {
		return "", "", err
	}

	// Get title of the frontmost window
	windowTitle, err := runAppleScript(fmt.Sprintf(`tell application "System Events" to tell process "%s" to get name of front window`, appName))
	if err != nil {
		return "", "", err
	}

	return appName, windowTitle, nil
}

// Get the URL of the active tab in the browser
func getBrowserUrl(browser string) (string, error) {
	script, ok := browserUrlScripts[browser]
	if !ok {
		return "", nil
	}

	return runAppleScript(script)
}

// Get the domain of a URL
func getDomain(urlString string) string {
	if urlString == "" {
		return ""
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return ""
	}
	return parsedURL.Hostname()
}

// Get the domain of the active tab in the browser
func getBrowserDomain(browser string) (string, error) {
	url, err := getBrowserUrl(browser)
	if err != nil {
		return "", err
	}

	return getDomain(url), nil
}
