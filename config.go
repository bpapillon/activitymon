package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

type DatabaseConfig struct {
	Type            string `json:"type"` // "sqlite" or "postgres"
	PostgresConnStr string `json:"postgresConnStr"`
}

type Config struct {
	Database DatabaseConfig `json:"database"`
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to get home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, "Library", "Preferences", "activitymon")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("unable to create config directory: %v", err)
	}

	return configDir, nil
}

func getConfigPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "json"), nil
}

func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	config := &Config{
		Database: DatabaseConfig{
			Type: "sqlite",
		},
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %v", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("unable to parse config file: %v", err)
	}

	return config, nil
}

func saveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal config: %v", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func configCmd() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Configure database settings",
		Subcommands: []*cli.Command{
			{
				Name:  "use-sqlite",
				Usage: "Use SQLite database (default)",
				Action: func(c *cli.Context) error {
					cfg, err := loadConfig()
					if err != nil {
						return err
					}
					cfg.Database.Type = "sqlite"
					return saveConfig(cfg)
				},
			},
			{
				Name:  "use-postgres",
				Usage: "Use PostgreSQL database",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "connection-string",
						Usage:    "PostgreSQL connection string",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					cfg, err := loadConfig()
					if err != nil {
						return err
					}
					cfg.Database.Type = "postgres"
					cfg.Database.PostgresConnStr = c.String("connection-string")
					return saveConfig(cfg)
				},
			},
		},
	}
}
