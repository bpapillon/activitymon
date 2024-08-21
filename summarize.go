package main

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

const (
	MaxGap            = 10 * time.Minute
	MaxActivityLength = 50
)

type Activity struct {
	Name     string
	Duration time.Duration
}

func getDomain(urlString string) string {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "Unknown"
	}
	return parsedURL.Hostname()
}

func truncateString(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength-3] + "..."
	}
	return s
}

func processActivities(results [][]interface{}) (map[string]time.Duration, time.Duration, [][]interface{}) {
	activitySummary := make(map[string]time.Duration)
	var totalDuration time.Duration
	var gaps [][]interface{}
	var lastTimestamp time.Time
	var lastActivity string

	for _, row := range results {
		timestamp := row[1].(time.Time)
		appName := row[2].(string)
		url := row[4].(string)

		activity := url
		if appName != "Brave Browser" && appName != "Google Chrome" && appName != "Safari" && appName != "Firefox" {
			activity = appName
		}
		if strings.HasPrefix(activity, "http") {
			activity = getDomain(activity)
		}

		if !lastTimestamp.IsZero() && lastActivity != "" {
			duration := timestamp.Sub(lastTimestamp)
			if duration > MaxGap || lastActivity == "loginwindow" {
				gaps = append(gaps, []interface{}{lastTimestamp, timestamp, duration})
			} else if lastActivity != "loginwindow" {
				activitySummary[lastActivity] += duration
				totalDuration += duration
			}
		}

		lastTimestamp = timestamp
		lastActivity = activity
	}

	return activitySummary, totalDuration, gaps
}

func formatTime(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dh %dm %ds", h, m, s)
}

func summarize(c *cli.Context) error {
	hours := c.Int("hours")
	minutes := c.Int("minutes")

	if hours == 0 && minutes == 0 {
		return fmt.Errorf("please specify either --hours or --minutes")
	}
	if hours != 0 && minutes != 0 {
		return fmt.Errorf("please specify either --hours or --minutes, not both")
	}

	var timeDelta time.Duration
	var timeUnit string
	var timeValue int

	if hours != 0 {
		timeDelta = time.Duration(hours) * time.Hour
		timeUnit = "hour(s)"
		timeValue = hours
	} else {
		timeDelta = time.Duration(minutes) * time.Minute
		timeUnit = "minute(s)"
		timeValue = minutes
	}

	db, err := getDb()
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer db.Close()

	now := time.Now()
	startTime := now.Add(-timeDelta)

	query := `
	SELECT *
	FROM activities
	WHERE timestamp >= ? AND timestamp < ?
	ORDER BY timestamp
	`

	rows, err := db.Query(query, startTime.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))
	if err != nil {
		return fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	var results [][]interface{}
	for rows.Next() {
		var id int
		var timestamp time.Time
		var appName, windowTitle, url string
		err := rows.Scan(&id, &timestamp, &appName, &windowTitle, &url)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}
		results = append(results, []interface{}{id, timestamp, appName, windowTitle, url})
	}

	activitySummary, totalDuration, gaps := processActivities(results)

	fmt.Printf("Activity Summary for the last %d %s\n\n", timeValue, timeUnit)

	if len(results) > 0 {
		fmt.Printf("Data range: %s to %s\n", results[0][1].(time.Time).Format("2006-01-02 15:04:05"), results[len(results)-1][1].(time.Time).Format("2006-01-02 15:04:05"))
	}

	requestedDuration := timeDelta
	totalGapTime := requestedDuration - totalDuration

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Metric", "Duration"})
	table.Append([]string{"Total tracked time", formatTime(totalDuration)})
	table.Append([]string{"Requested duration", timeDelta.String()})
	table.Append([]string{"Total gap time", formatTime(totalGapTime)})
	table.Render()

	coveragePercentage := float64(totalDuration) / float64(requestedDuration) * 100
	fmt.Printf("\nTracking coverage: %.2f%%\n", coveragePercentage)

	if len(gaps) > 0 {
		fmt.Printf("\nDetected gaps within tracked time: %d\n", len(gaps))
		fmt.Println("Largest gaps within tracked time:")
		sort.Slice(gaps, func(i, j int) bool {
			return gaps[i][2].(time.Duration) > gaps[j][2].(time.Duration)
		})
		for i, gap := range gaps[:min(5, len(gaps))] {
			fmt.Printf("  %d. From %s to %s (%s)\n", i+1, gap[0].(time.Time).Format("2006-01-02 15:04:05"), gap[1].(time.Time).Format("2006-01-02 15:04:05"), formatTime(gap[2].(time.Duration)))
		}
	}

	fmt.Println("\nTop activities (% of tracked time, excluding sleep):")
	activityTable := tablewriter.NewWriter(os.Stdout)
	activityTable.SetHeader([]string{"Activity", "Duration", "Percentage"})

	var activities []Activity
	for activity, duration := range activitySummary {
		if activity != "loginwindow" {
			activities = append(activities, Activity{Name: activity, Duration: duration})
		}
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Duration > activities[j].Duration
	})

	for _, activity := range activities {
		percentage := float64(activity.Duration) / float64(totalDuration) * 100
		if percentage > 0.5 {
			activityTable.Append([]string{
				truncateString(activity.Name, MaxActivityLength),
				formatTime(activity.Duration),
				fmt.Sprintf("%.2f%%", percentage),
			})
		}
	}
	activityTable.Render()

	sleepTime := time.Duration(0)
	for _, gap := range gaps {
		if gap[2].(time.Duration) > MaxGap {
			sleepTime += gap[2].(time.Duration)
		}
	}
	fmt.Printf("\nNote: Your device was likely asleep or locked for approximately %s.\n", formatTime(sleepTime))

	return nil
}
