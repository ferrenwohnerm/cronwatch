package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job describes a single monitored cron job.
type Job struct {
	Name             string        `yaml:"name"`
	ExpectedDuration time.Duration `yaml:"expected_duration"`
	Tolerance        time.Duration `yaml:"tolerance"`
}

// Config holds the full daemon configuration.
type Config struct {
	PollInterval time.Duration `yaml:"poll_interval"`
	Jobs         []Job         `yaml:"jobs"`
	WebhookURL   string        `yaml:"webhook_url"`
}

const (
	defaultPollInterval = 30 * time.Second
	defaultTolerance    = 5 * time.Second
)

// Load reads and validates a YAML config file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = defaultPollInterval
	}
	for i := range cfg.Jobs {
		if cfg.Jobs[i].Name == "" {
			return nil, errors.New("job missing required field: name")
		}
		if cfg.Jobs[i].ExpectedDuration <= 0 {
			return nil, errors.New("job " + cfg.Jobs[i].Name + ": expected_duration must be positive")
		}
		if cfg.Jobs[i].Tolerance <= 0 {
			cfg.Jobs[i].Tolerance = defaultTolerance
		}
	}
	return &cfg, nil
}
