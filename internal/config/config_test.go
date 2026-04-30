package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	raw := `
check_interval: 2m
log_level: debug
jobs:
  - name: nightly-backup
    schedule: "0 2 * * *"
    expected_start: "02:00"
    window: 10m
    alert_email: ops@example.com
`
	path := writeTempConfig(t, raw)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 2*time.Minute {
		t.Errorf("expected check_interval 2m, got %v", cfg.CheckInterval)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "nightly-backup" {
		t.Errorf("unexpected job name: %q", cfg.Jobs[0].Name)
	}
	if cfg.Jobs[0].Window != 10*time.Minute {
		t.Errorf("expected window 10m, got %v", cfg.Jobs[0].Window)
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	raw := `
jobs:
  - name: sync-job
    schedule: "*/5 * * * *"
`
	path := writeTempConfig(t, raw)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 1*time.Minute {
		t.Errorf("expected default check_interval 1m, got %v", cfg.CheckInterval)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log_level info, got %q", cfg.LogLevel)
	}
	if cfg.Jobs[0].Window != 5*time.Minute {
		t.Errorf("expected default window 5m, got %v", cfg.Jobs[0].Window)
	}
}

func TestLoad_MissingJobName(t *testing.T) {
	raw := `
jobs:
  - schedule: "0 3 * * *"
`
	path := writeTempConfig(t, raw)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing job name, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
