package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Browser-specific AppleScript snippets to get the URL of the active tab in the browser
var browserUrlScripts = map[string]string{
	"Arc":    "tell application \"Arc\" to set currentURL to URL of active tab of front window",
	"Safari": "tell application \"Safari\" to set currentURL to URL of current tab of window 1",
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

// Get the name of the frontmost application and the title of the frontmost window
func getAppAndWindow() (string, string, error) {
	appName, err := runAppleScript(`tell application "System Events" to get name of first process whose frontmost is true`)
	if err != nil {
		return "", "", err
	}

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
