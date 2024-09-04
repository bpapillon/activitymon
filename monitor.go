package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
)

func monitorCmd(c *cli.Context) error {
	db, err := getDb()
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	// Clean up any unfinished activities from previous runs
	if err := cleanupUnfinishedActivities(db); err != nil {
		fmt.Printf("Error cleaning up unfinished activities: %v\n", err)
	}

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start the monitor in a goroutine
	errChan := make(chan error)
	go func() {
		errChan <- monitor(ctx, db)
	}()

	// Wait for either an error from monitor or a signal
	select {
	case err := <-errChan:
		return err
	case sig := <-signalChan:
		fmt.Printf("Received signal: %v\n", sig)
		cancel()
		if err := endCurrentActivity(db, time.Now()); err != nil {
			fmt.Printf("Error ending current activity: %v\n", err)
		}
		return nil
	}
}

func monitor(ctx context.Context, db *sql.DB) error {
	var lastAppName, lastDomain string

	for {
		currentTime := time.Now()
		appName, windowTitle, err := getAppAndWindow()
		if err != nil {
			fmt.Printf("Failed to get window info: %v. Retrying in 1 second...\n", err)
			time.Sleep(time.Second)
			continue
		}
		domain, err := getBrowserDomain(appName)
		if err != nil {
			fmt.Printf("Failed to get browser URL info: %s\n", err)
		}

		if appName == "" && windowTitle == "" {
			// Computer is likely asleep or locked
			if lastAppName != "" {
				if err := endCurrentActivity(db, currentTime); err != nil {
					fmt.Println("Error ending current activity:", err)
				}
				lastAppName = ""
				lastDomain = ""
			}
		} else if appName != lastAppName || domain != lastDomain {
			// Activity has changed
			if lastAppName != "" {
				if err := endCurrentActivity(db, currentTime); err != nil {
					fmt.Println("Error ending current activity:", err)
				}
			}

			activityName := appName
			if domain != "" {
				activityName = domain
			}
			if err := insertActivity(db, currentTime, activityName); err != nil {
				fmt.Println("Error inserting activity:", err)
			} else {
				fmt.Printf("Started activity: %s\n", activityName)
			}

			lastAppName = appName
			lastDomain = domain
		}

		time.Sleep(1 * time.Second)
	}
}
