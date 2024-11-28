package main

import (
	"context"
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

	if err := db.cleanupUnfinishedActivities(); err != nil {
		fmt.Printf("Error cleaning up unfinished activities: %v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error)
	go func() {
		errChan <- monitor(ctx, db)
	}()

	select {
	case err := <-errChan:
		return err
	case sig := <-signalChan:
		fmt.Printf("Received signal: %v\n", sig)
		cancel()
		if err := db.endCurrentActivity(time.Now()); err != nil {
			fmt.Printf("Error ending current activity: %v\n", err)
		}
		return nil
	}
}

func monitor(ctx context.Context, db *DB) error {
	display := NewMonitor()
	if err := display.Start(); err != nil {
		return err
	}
	defer display.Stop()

	var lastAppName, lastDomain string
	ticker := time.NewTicker(time.Second)
	statsTicker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			currentTime := time.Now()
			appName, windowTitle, err := getAppAndWindow()
			if err != nil {
				display.AddLogEntry(fmt.Sprintf("[red]Failed to get window info: %v. Retrying...[white]", err))
				if err := db.endCurrentActivity(currentTime); err != nil {
					display.AddLogEntry(fmt.Sprintf("[red]Error ending current activity: %v[white]", err))
				}
				lastAppName = ""
				lastDomain = ""
				continue
			}

			domain, err := getBrowserDomain(appName)
			if err != nil {
				display.AddLogEntry(fmt.Sprintf("[red]Failed to get browser URL info: %v[white]", err))
			}

			if appName == "" && windowTitle == "" {
				// computer is likely asleep or locked
				if lastAppName != "" {
					if err := db.endCurrentActivity(currentTime); err != nil {
						display.AddLogEntry(fmt.Sprintf("[red]Error ending current activity: %v[white]", err))
					}
					lastAppName = ""
					lastDomain = ""
				}
			} else if appName != lastAppName || domain != lastDomain {
				// activity has changed
				if lastAppName != "" {
					if err := db.endCurrentActivity(currentTime); err != nil {
						display.AddLogEntry(fmt.Sprintf("[red]Error ending current activity: %v[white]", err))
					}
				}

				activityName := appName
				if domain != "" {
					activityName = domain
				}
				if err := db.insertActivity(currentTime, activityName); err != nil {
					display.AddLogEntry(fmt.Sprintf("[red]Error inserting activity: %v[white]", err))
				} else {
					display.AddLogEntry(fmt.Sprintf("Started activity: %s", activityName))
				}

				lastAppName = appName
				lastDomain = domain
				if appName != lastAppName || domain != lastDomain {
					display.AddLogEntry(fmt.Sprintf("[green]%s Started activity: %s[white]",
						currentTime.Format("2006-01-02 15:04:05"),
						activityName))
				}
			}

		case <-statsTicker.C:
			stats, err := getLatestStats(db)
			if err != nil {
				display.AddLogEntry(fmt.Sprintf("[red]Error updating stats: %v[white]", err))
				continue
			}
			display.UpdateStats(stats)
		}
	}
}
