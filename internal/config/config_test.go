package config_test

import (
	"testing"
	"time"

	"github.com/erkexzcx/justimport/internal/config"
)

// clearEnv unsets all config-related environment variables for the duration of the test.
func clearEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"RADARR_URL", "RADARR_API_KEY",
		"SONARR_URL", "SONARR_API_KEY",
		"POLL_INTERVAL", "IMPORT_MODE", "DRY_RUN",
	} {
		t.Setenv(key, "")
	}
}

func TestLoad_NeitherConfigured(t *testing.T) {
	clearEnv(t)

	_, err := config.Load()
	if err == nil {
		t.Error("expected error when neither Radarr nor Sonarr is configured")
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("RADARR_URL", "http://radarr:7878")
	t.Setenv("RADARR_API_KEY", "radarr-key")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 60*time.Second {
		t.Errorf("expected 60s default interval, got %v", cfg.PollInterval)
	}
	if cfg.ImportMode != "Move" {
		t.Errorf("expected Move default import mode, got %s", cfg.ImportMode)
	}
	if !cfg.DryRun {
		t.Error("expected DryRun=true by default")
	}
}

func TestLoad_OnlyRadarr(t *testing.T) {
	clearEnv(t)
	t.Setenv("RADARR_URL", "http://radarr:7878")
	t.Setenv("RADARR_API_KEY", "radarr-key")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RadarrURL != "http://radarr:7878" {
		t.Errorf("unexpected RadarrURL: %s", cfg.RadarrURL)
	}
	if cfg.SonarrURL != "" {
		t.Errorf("expected SonarrURL to be empty, got %s", cfg.SonarrURL)
	}
}

func TestLoad_OnlySonarr(t *testing.T) {
	clearEnv(t)
	t.Setenv("SONARR_URL", "http://sonarr:8989")
	t.Setenv("SONARR_API_KEY", "sonarr-key")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.SonarrURL != "http://sonarr:8989" {
		t.Errorf("unexpected SonarrURL: %s", cfg.SonarrURL)
	}
	if cfg.RadarrURL != "" {
		t.Errorf("expected RadarrURL to be empty, got %s", cfg.RadarrURL)
	}
}

func TestLoad_BothConfigured(t *testing.T) {
	clearEnv(t)
	t.Setenv("RADARR_URL", "http://radarr:7878")
	t.Setenv("RADARR_API_KEY", "radarr-key")
	t.Setenv("SONARR_URL", "http://sonarr:8989")
	t.Setenv("SONARR_API_KEY", "sonarr-key")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RadarrURL == "" || cfg.SonarrURL == "" {
		t.Error("expected both Radarr and Sonarr to be configured")
	}
}

func TestLoad_CustomPollInterval(t *testing.T) {
	clearEnv(t)
	t.Setenv("RADARR_URL", "http://radarr:7878")
	t.Setenv("POLL_INTERVAL", "5m")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 5*time.Minute {
		t.Errorf("expected 5m, got %v", cfg.PollInterval)
	}
}

func TestLoad_InvalidPollInterval(t *testing.T) {
	clearEnv(t)
	t.Setenv("RADARR_URL", "http://radarr:7878")
	t.Setenv("POLL_INTERVAL", "not-a-duration")

	_, err := config.Load()
	if err == nil {
		t.Error("expected error for invalid POLL_INTERVAL")
	}
}

func TestLoad_DryRunFalse(t *testing.T) {
	clearEnv(t)
	t.Setenv("RADARR_URL", "http://radarr:7878")
	t.Setenv("DRY_RUN", "false")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DryRun {
		t.Error("expected DryRun=false")
	}
}

func TestLoad_DryRunFalseUppercase(t *testing.T) {
	clearEnv(t)
	t.Setenv("RADARR_URL", "http://radarr:7878")
	t.Setenv("DRY_RUN", "FALSE")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DryRun {
		t.Error("expected DryRun=false for DRY_RUN=FALSE")
	}
}

func TestLoad_DryRunTrueExplicit(t *testing.T) {
	clearEnv(t)
	t.Setenv("RADARR_URL", "http://radarr:7878")
	t.Setenv("DRY_RUN", "true")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.DryRun {
		t.Error("expected DryRun=true for DRY_RUN=true")
	}
}

func TestLoad_CustomImportMode(t *testing.T) {
	clearEnv(t)
	t.Setenv("RADARR_URL", "http://radarr:7878")
	t.Setenv("IMPORT_MODE", "Copy")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ImportMode != "Copy" {
		t.Errorf("expected Copy, got %s", cfg.ImportMode)
	}
}
