package main

import (
	"database/sql"
	"fmt"
	"time"
)

const dbName = "tracker.db"

func cleanupUnfinishedActivities(db *sql.DB) error {
	now := time.Now()
	fiveMinutesAgo := now.Add(-5 * time.Minute)

	_, err := db.Exec(`
		UPDATE activities
		SET end_time = CASE
			WHEN start_time > ? THEN start_time
			ELSE ?
		END
		WHERE end_time IS NULL
	`, fiveMinutesAgo.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))

	return err
}

func endCurrentActivity(db *sql.DB, endTime time.Time) error {
	_, err := db.Exec(`
		UPDATE activities
		SET end_time = ?
		WHERE end_time IS NULL
	`, endTime.Format("2006-01-02 15:04:05"))

	return err
}

func getDb() (*sql.DB, error) {
	if err := setupDatabase(); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func insertActivity(db *sql.DB, startTime time.Time, activityName string) error {
	_, err := db.Exec(`
		INSERT INTO activities (start_time, activity_name)
		VALUES (?, ?)
	`, startTime.Format("2006-01-02 15:04:05"), activityName)

	return err
}

func setupDatabase() error {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS activities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		activity_name TEXT NOT NULL
	)
	`); err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	if _, err := db.Exec(`
    CREATE INDEX IF NOT EXISTS idx_activities_start_time ON activities(start_time)
    `); err != nil {
		return fmt.Errorf("error creating index: %v", err)
	}

	return nil
}
