package main

import (
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "activitymon",
		Usage: "Simple activity tracker for Mac OS",
		Commands: []*cli.Command{
			{
				Name:   "monitor",
				Usage:  "Run the activity monitor",
				Action: monitor,
			},
			{
				Name:  "summary",
				Usage: "Output a summary of activities for the last N hours or minutes",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "hours",
						Usage: "Number of hours to summarize",
					},
					&cli.IntFlag{
						Name:  "minutes",
						Usage: "Number of minutes to summarize",
					},
				},
				Action: summarize,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
