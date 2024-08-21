package main

import (
	"database/sql"
	"fmt"
)

const dbName = "tracker.db"

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

func insertActivity(db *sql.DB, timestamp, appName, windowTitle, url string) error {
	_, err := db.Exec(`
		INSERT INTO activities (timestamp, app_name, window_title, url)
		VALUES (?, ?, ?, ?)
	`, timestamp, appName, windowTitle, url)
	return err
}

func setupDatabase() error {
	// if _, err := os.Stat(dbName); err != nil {
	// 	fmt.Println("Database exists")
	// }

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS activities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		app_name TEXT NOT NULL,
		window_title TEXT,
		url TEXT
	)
	`)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	return nil
}
