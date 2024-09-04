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
				Action: monitorCmd,
			},
			{
				Name:  "summary",
				Usage: "Output a summary of activities",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "minutes",
						Usage: "Number of minutes to summarize",
						Value: 240,
					},
				},
				Action: summaryCmd,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
