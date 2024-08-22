package main

import (
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
)

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
		appName, windowTitle, err := getAppAndWindow()
		if err != nil {
			fmt.Printf("Failed to get window info: %v. Retrying in 1 second...\n", err)
			time.Sleep(time.Second)
			continue
		}
		url, err := getBrowserUrl(appName)
		if err != nil {
			fmt.Printf("Failed to get URL info: %s\n", err)
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
