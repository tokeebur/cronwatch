package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job represents a single monitored cron job.
type Job struct {
	Name    string        `yaml:"name"`
	Command string        `yaml:"command"`
	Timeout time.Duration `yaml:"timeout"`
}

// Alerting holds notification settings.
type Alerting struct {
	Email   string `yaml:"email"`
	Webhook string `yaml:"webhook"`
}

// Config is the top-level configuration structure.
type Config struct {
	LogLevel string   `yaml:"log_level"`
	Alerting Alerting `yaml:"alerting"`
	Jobs     []Job    `yaml:"jobs"`
}

// Load reads and parses the YAML configuration file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if len(cfg.Jobs) == 0 {
		return errors.New("config: at least one job must be defined")
	}
	for i, j := range cfg.Jobs {
		if j.Command == "" {
			return fmt.Errorf("config: job[%d] %q missing command", i, j.Name)
		}
	}
	return nil
}
