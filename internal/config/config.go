package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds the application configuration parsed from environment variables.
type Config struct {
	RadarrURL    string
	RadarrAPIKey string
	SonarrURL    string
	SonarrAPIKey string
	PollInterval time.Duration
	ImportMode   string
	DryRun       bool
}

// Load reads and validates configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		RadarrURL:    os.Getenv("RADARR_URL"),
		RadarrAPIKey: os.Getenv("RADARR_API_KEY"),
		SonarrURL:    os.Getenv("SONARR_URL"),
		SonarrAPIKey: os.Getenv("SONARR_API_KEY"),
	}

	if cfg.RadarrURL == "" && cfg.SonarrURL == "" {
		return nil, errors.New("at least one of RADARR_URL or SONARR_URL must be set")
	}

	pollIntervalStr := os.Getenv("POLL_INTERVAL")
	if pollIntervalStr == "" {
		pollIntervalStr = "60s"
	}

	dur, err := time.ParseDuration(pollIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid POLL_INTERVAL %q: %w", pollIntervalStr, err)
	}

	cfg.PollInterval = dur

	importMode := os.Getenv("IMPORT_MODE")
	if importMode == "" {
		importMode = "Move"
	}

	cfg.ImportMode = importMode

	dryRunStr := os.Getenv("DRY_RUN")
	if dryRunStr == "" {
		cfg.DryRun = true
	} else {
		cfg.DryRun = strings.ToLower(dryRunStr) != "false"
	}

	return cfg, nil
}
