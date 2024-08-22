package main

import (
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

char* getActiveAppName() {
    @autoreleasepool {
        NSWorkspace *workspace = [NSWorkspace sharedWorkspace];
        NSRunningApplication *runningApp = [workspace frontmostApplication];
        return strdup([[runningApp localizedName] UTF8String]);
    }
}

char* getActiveWindowTitle(const char* appName) {
    @autoreleasepool {
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
            return strdup([[result stringValue] UTF8String]);
        } else {
            return strdup("No window title available");
        }
    }
}
*/
import "C"

func getActiveWindowInfo() (string, string, string, error) {
	appName := C.GoString(C.getActiveAppName())
	windowTitle := C.GoString(C.getActiveWindowTitle(C.CString(appName)))

	url, err := getBrowserUrl(appName)

	return appName, windowTitle, url, err
}

func monitor(c *cli.Context) error {
	db, err := getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	var lastAppName, lastWindowTitle, lastURL string
	var lastInsertTime time.Time

	time.Sleep(100 * time.Millisecond)

	for {
		currentTime := time.Now().Format("2006-01-02 15:04:05")
		appName, windowTitle, url, err := getActiveWindowInfo()
		if err != nil {
			fmt.Printf("Failed to get window info: %v. Retrying in 1 second...\n", err)
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
		} else {
			fmt.Println("No change detected")
		}

		time.Sleep(1000 * time.Millisecond)
	}
}
