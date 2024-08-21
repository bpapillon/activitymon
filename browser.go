package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func getBrowserUrl(browser string) (string, error) {
	script, ok := browserUrlScripts[browser]
	if !ok {
		return "", nil
	}

	cmd := exec.Command("osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("AppleScript error: %v, stderr: %s", err, stderr.String())
	}
	url := strings.TrimSpace(stdout.String())
	return url, nil
}

var browserUrlScripts = map[string]string{
	"Safari": `tell application "Safari"
	if not (exists window 1) then
		return "No window found"
	end if

	tell window 1
		set tabURL to URL of current tab
	end tell

	return tabURL
end tell`,
	"Arc": "tell application \"Arc\" to set currentURL to URL of active tab of front window",
}
