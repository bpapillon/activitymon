package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

char* getActiveAppName() {
    NSWorkspace *workspace = [NSWorkspace sharedWorkspace];
    NSRunningApplication *runningApp = [workspace frontmostApplication];
    return (char*)[[runningApp localizedName] UTF8String];
}

char* getActiveWindowTitle(const char* appName) {
    NSString *appNameStr = [NSString stringWithUTF8String:appName];
    NSString *script = [NSString stringWithFormat:@"tell application \"System Events\"\n"
                        "    tell process \"%@\"\n"
                        "        try\n"
                        "            return name of front window\n"
                        "        on error\n"
                        "            return \"No window title available\"\n"
                        "        end try\n"
                        "    end tell\n"
                        "end tell", appNameStr];

    NSDictionary *error = nil;
    NSAppleScript *appleScript = [[NSAppleScript alloc] initWithSource:script];
    NSAppleEventDescriptor *result = [appleScript executeAndReturnError:&error];

    if (result != nil) {
        return (char*)[[result stringValue] UTF8String];
    } else {
        return "No window title available";
    }
}
*/
import "C"

func getActiveWindowInfo() (string, string, string, error) {
	appName := C.GoString(C.getActiveAppName())
	windowTitle := C.GoString(C.getActiveWindowTitle(C.CString(appName)))

	var url string
	if appName == "Safari" || appName == "Google Chrome" || appName == "Firefox" || appName == "Brave Browser" || appName == "Arc" {
		script := fmt.Sprintf(`
			tell application "%s"
				set currentTab to active tab of front window
				return (URL of currentTab) & "|" & (name of currentTab)
			end tell
		`, appName)
		result, err := runAppleScript(script)
		if err == nil {
			fmt.Println("result", result)
			parts := strings.SplitN(result, "|", 2)
			if len(parts) == 2 {
				url = strings.TrimSpace(parts[0])
				windowTitle = strings.TrimSpace(parts[1])
			}
		} else {
			fmt.Printf("Error getting URL from browser: %s\n", err)
		}
	}

	return appName, windowTitle, url, nil
}

func runAppleScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func monitor(c *cli.Context) error {
	// Set up the database if it doesn't exist
	if err := setupDatabase(); err != nil {
		return err
	}

	db, err := getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	var lastAppName, lastWindowTitle, lastURL string
	var lastInsertTime time.Time

	for {
		currentTime := time.Now().Format("2006-01-02 15:04:05")
		appName, windowTitle, url, err := getActiveWindowInfo()
		if err != nil {
			fmt.Println("Failed to get window info. Retrying in 1 second...")
			time.Sleep(time.Second)
			continue
		}

		if appName != lastAppName ||
			windowTitle != lastWindowTitle ||
			url != lastURL ||
			time.Since(lastInsertTime) >= time.Second {

			if err := insertActivity(db, currentTime, appName, windowTitle, url); err != nil {
				fmt.Println("Error inserting activity:", err)
			} else {
				fmt.Printf("Inserted %s, %s, %s, %s\n", currentTime, appName, windowTitle, url)
			}

			lastAppName = appName
			lastWindowTitle = windowTitle
			lastURL = url
			lastInsertTime = time.Now()
		}

		time.Sleep(time.Second)
	}
}
