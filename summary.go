package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
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

type SummaryData struct {
	Activities    []Activity
	TotalDuration time.Duration
	TimePeriod    time.Duration
}

// getSummaryData retrieves activity data from the database
func getSummaryData(db *DB, minutes int) (*SummaryData, error) {
	timeDelta := time.Duration(minutes) * time.Minute
	now := time.Now()
	startTime := now.Add(-timeDelta)

	var query string
	if db.dbType == "postgres" {
		query = `
		SELECT
			activity_name,
			EXTRACT(EPOCH FROM (
				CASE
					WHEN COALESCE(end_time, CURRENT_TIMESTAMP) > $1 THEN $1
					ELSE COALESCE(end_time, CURRENT_TIMESTAMP)
				END
				-
				CASE
					WHEN start_time < $3 THEN $3
					ELSE start_time
				END
			)) AS duration_seconds
		FROM activities
		WHERE start_time < $5 AND (end_time > $6 OR end_time IS NULL)
		GROUP BY activity_name
		ORDER BY duration_seconds DESC`
	} else {
		query = `
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
		ORDER BY duration_seconds DESC`
	}

	rows, err := db.Query(query,
		now.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"),
		startTime.Format("2006-01-02 15:04:05"), startTime.Format("2006-01-02 15:04:05"),
		now.Format("2006-01-02 15:04:05"), startTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	var activities []Activity
	var totalDuration time.Duration

	for rows.Next() {
		var activity string
		var durationSeconds float64
		if err := rows.Scan(&activity, &durationSeconds); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		duration := time.Duration(durationSeconds) * time.Second
		if duration > timeDelta {
			duration = timeDelta
		}
		activities = append(activities, Activity{Name: activity, Duration: duration})
		totalDuration += duration
	}

	return &SummaryData{
		Activities:    activities,
		TotalDuration: totalDuration,
		TimePeriod:    timeDelta,
	}, nil
}

func formatSummary(data *SummaryData) string {
	if len(data.Activities) == 0 {
		return "[yellow]No activity data found for the last %d minutes[white]\n"
	}

	var buf strings.Builder

	// Create header
	buf.WriteString("[cyan]🕒 Activity Summary for the last " + fmt.Sprintf("%d", int(data.TimePeriod.Minutes())) + " minutes[white]\n\n")

	// Create simple table header
	buf.WriteString("[cyan]╔══════════════════════╦═══════════════════╗[white]\n")
	buf.WriteString("[cyan]║[yellow] Total tracked time [cyan]║[green] " +
		fmt.Sprintf("%-17s", formatTime(data.TotalDuration)) + "[cyan]║[white]\n")
	buf.WriteString("[cyan]╚══════════════════════╩═══════════════════╝[white]\n\n")

	// Top activities header
	buf.WriteString("[cyan]🏆 Top activities (% of tracked time):[white]\n")

	// Sort activities by duration
	sort.Slice(data.Activities, func(i, j int) bool {
		return data.Activities[i].Duration > data.Activities[j].Duration
	})

	// Create activity bars
	maxDuration := data.Activities[0].Duration
	for _, activity := range data.Activities {
		percentage := float64(activity.Duration) / float64(data.TotalDuration) * 100
		if percentage > 0.5 {
			barLength := int(float64(activity.Duration) / float64(maxDuration) * BarChartWidth)
			bar := strings.Repeat("█", barLength) + strings.Repeat("░", BarChartWidth-barLength)

			// Format each line with tview color tags
			buf.WriteString(fmt.Sprintf("[lightblue]%-30s[yellow] %s [green]%5.2f%% [magenta]%s[white]\n",
				truncateString(activity.Name, 30),
				formatTime(activity.Duration),
				percentage,
				bar))
		}
	}

	return buf.String()
}

func summaryCmd(c *cli.Context) error {
	db, err := getDb()
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer db.Close()

	data, err := getSummaryData(db, c.Int("minutes"))
	if err != nil {
		return err
	}

	fmt.Print(formatSummary(data))
	return nil
}

func getLatestStats(db *DB) (string, error) {
	data, err := getSummaryData(db, 240) // Default to last 4 hours
	if err != nil {
		return "", err
	}
	return formatSummary(data), nil
}
