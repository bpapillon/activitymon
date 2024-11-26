package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

const (
	MaxActivityLength = 50
	BarChartWidth     = 50
)

type Activity struct {
	Name     string
	Duration time.Duration
}

func summaryCmd(c *cli.Context) error {
	db, err := getDb()
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer db.Close()

	minutes := c.Int("minutes")
	timeDelta := time.Duration(minutes) * time.Minute
	now := time.Now()
	startTime := now.Add(-timeDelta)

	query := `
	SELECT
    activity_name,
		SUM(
			JULIANDAY(
				CASE
					WHEN COALESCE(end_time, CURRENT_TIMESTAMP) > ? THEN ?
					ELSE COALESCE(end_time, CURRENT_TIMESTAMP)
				END
			) - JULIANDAY(
				CASE
					WHEN start_time < ? THEN ?
					ELSE start_time
				END
			)
		) * 86400 AS duration_seconds
	FROM activities
	WHERE start_time < ? AND (end_time > ? OR end_time IS NULL)
	GROUP BY activity_name
	ORDER BY duration_seconds DESC
	`

	rows, err := db.Query(query,
		now.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"),
		startTime.Format("2006-01-02 15:04:05"), startTime.Format("2006-01-02 15:04:05"),
		now.Format("2006-01-02 15:04:05"), startTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		return fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	var activities []Activity
	var totalDuration time.Duration

	for rows.Next() {
		var activity string
		var durationSeconds float64
		if err := rows.Scan(&activity, &durationSeconds); err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}
		duration := time.Duration(durationSeconds) * time.Second
		if duration > timeDelta {
			duration = timeDelta
		}
		activities = append(activities, Activity{Name: activity, Duration: duration})
		totalDuration += duration
	}

	if len(activities) == 0 {
		color.Yellow("No activity data found for the last %d minutes\n", minutes)
		return nil
	}

	color.Cyan("🕒 Activity Summary for the last %d minutes\n\n", minutes)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Metric", "Duration"})
	table.SetBorder(false)
	table.SetColumnSeparator("│")
	table.SetCenterSeparator("─")
	table.SetHeaderColor(tablewriter.Colors{tablewriter.FgHiCyanColor}, tablewriter.Colors{tablewriter.FgHiCyanColor})
	table.SetColumnColor(tablewriter.Colors{tablewriter.FgYellowColor}, tablewriter.Colors{tablewriter.FgGreenColor})

	table.Append([]string{"Total tracked time", formatTime(totalDuration)})
	table.Render()

	color.Cyan("\n🏆 Top activities (% of tracked time):\n")

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Duration > activities[j].Duration
	})

	maxDuration := activities[0].Duration
	for _, activity := range activities {
		percentage := float64(activity.Duration) / float64(totalDuration) * 100
		if percentage > 0.5 {
			barLength := int(float64(activity.Duration) / float64(maxDuration) * BarChartWidth)
			bar := strings.Repeat("█", barLength) + strings.Repeat("░", BarChartWidth-barLength)
			color.Set(color.FgHiBlue)
			fmt.Printf("%-30s", truncateString(activity.Name, 30))
			color.Set(color.FgHiYellow)
			fmt.Printf(" %s ", formatTime(activity.Duration))
			color.Set(color.FgHiGreen)
			fmt.Printf("%5.2f%% ", percentage)
			color.Set(color.FgHiMagenta)
			fmt.Println(bar)
			color.Unset()
		}
	}

	return nil
}
