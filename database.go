package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
	dbType string
}

func getDb() (*DB, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	var db *sql.DB
	switch cfg.Database.Type {
	case "sqlite":
		configDir, err := getConfigDir()
		if err != nil {
			return nil, err
		}
		dbPath := filepath.Join(configDir, "tracker.db")
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			return nil, fmt.Errorf("error opening sqlite database: %v", err)
		}

	case "postgres":
		db, err = sql.Open("postgres", cfg.Database.PostgresConnStr)
		if err != nil {
			return nil, fmt.Errorf("error opening postgres database: %v", err)
		}

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}

	wrappedDB := &DB{db, cfg.Database.Type}

	// try to create tables if they don't exist
	if err := wrappedDB.Setup(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting up database: %v", err)
	}

	return wrappedDB, nil
}

func (db *DB) Setup() error {
	var createTable string
	if db.dbType == "postgres" {
		createTable = `
		CREATE TABLE IF NOT EXISTS activities (
			id SERIAL PRIMARY KEY,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP,
			activity_name TEXT NOT NULL
		)`
	} else {
		createTable = `
		CREATE TABLE IF NOT EXISTS activities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			activity_name TEXT NOT NULL
		)`
	}

	if _, err := db.Exec(createTable); err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_activities_start_time ON activities(start_time)
	`); err != nil {
		return fmt.Errorf("error creating index: %v", err)
	}

	return nil
}

func (db *DB) cleanupUnfinishedActivities() error {
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

func (db *DB) endCurrentActivity(endTime time.Time) error {
	_, err := db.Exec(`
		UPDATE activities
		SET end_time = ?
		WHERE end_time IS NULL
	`, endTime.Format("2006-01-02 15:04:05"))

	return err
}

func (db *DB) insertActivity(startTime time.Time, activityName string) error {
	_, err := db.Exec(`
		INSERT INTO activities (start_time, activity_name)
		VALUES (?, ?)
	`, startTime.Format("2006-01-02 15:04:05"), activityName)

	return err
}
