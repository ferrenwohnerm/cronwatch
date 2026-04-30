package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// JobConfig defines the expected execution window for a single cron job.
type JobConfig struct {
	Name          string        `yaml:"name"`
	Schedule      string        `yaml:"schedule"`
	ExpectedStart string        `yaml:"expected_start"` // e.g. "02:00"
	Window        time.Duration `yaml:"window"`         // allowed drift
	AlertEmail    string        `yaml:"alert_email"`
}

// Config is the top-level configuration for cronwatch.
type Config struct {
	CheckInterval time.Duration `yaml:"check_interval"`
	LogLevel      string        `yaml:"log_level"`
	Jobs          []JobConfig   `yaml:"jobs"`
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate checks that required fields are present and sensible.
func (c *Config) validate() error {
	if c.CheckInterval <= 0 {
		c.CheckInterval = 1 * time.Minute
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	for i, job := range c.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if job.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", job.Name)
		}
		if job.Window <= 0 {
			c.Jobs[i].Window = 5 * time.Minute
		}
	}
	return nil
}
