package main

import (
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "Activity Tracker",
		Usage: "Personal Activity Tracker CLI",
		Commands: []*cli.Command{
			{
				Name:   "monitor",
				Usage:  "Run the activity monitor",
				Action: monitor,
			},
			{
				Name:  "summary",
				Usage: "Provide a summary of activities for the specified time period",
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
