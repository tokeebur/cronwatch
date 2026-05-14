package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Job describes a single cron job to monitor.
type Job struct {
	Name            string         `yaml:"name"`
	Schedule        string         `yaml:"schedule"`
	Command         string         `yaml:"command"`
	TimeoutSeconds  int            `yaml:"timeout_seconds"`
	Retry           *RetryConfig   `yaml:"retry,omitempty"`
	Throttle        *ThrottleConfig `yaml:"throttle,omitempty"`
	CircuitBreaker  *CBConfig      `yaml:"circuit_breaker,omitempty"`
}

// RetryConfig holds retry parameters for a job.
type RetryConfig struct {
	MaxAttempts    int `yaml:"max_attempts"`
	BackoffSeconds int `yaml:"backoff_seconds"`
}

// ThrottleConfig holds alert-throttle parameters for a job.
type ThrottleConfig struct {
	CooldownSeconds int `yaml:"cooldown_seconds"`
}

// CBConfig holds circuit-breaker parameters for a job.
type CBConfig struct {
	MaxFailures     int `yaml:"max_failures"`
	CooldownSeconds int `yaml:"cooldown_seconds"`
}

// Config is the top-level configuration structure.
type Config struct {
	Jobs []Job `yaml:"jobs"`

	SMTP struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		From     string `yaml:"from"`
		To       string `yaml:"to"`
	} `yaml:"smtp"`

	Webhook struct {
		URL     string            `yaml:"url"`
		Method  string            `yaml:"method"`
		Headers map[string]string `yaml:"headers"`
	} `yaml:"webhook"`

	API struct {
		Addr string `yaml:"addr"`
	} `yaml:"api"`
}

// Load reads and validates a YAML config file from path.
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

// JobByName returns the Job with the given name, or an error if not found.
func (c *Config) JobByName(name string) (*Job, error) {
	for i := range c.Jobs {
		if c.Jobs[i].Name == name {
			return &c.Jobs[i], nil
		}
	}
	return nil, fmt.Errorf("config: no job named %q", name)
}

func validate(cfg *Config) error {
	if len(cfg.Jobs) == 0 {
		return errors.New("config: no jobs defined")
	}
	for i, j := range cfg.Jobs {
		if j.Command == "" {
			return fmt.Errorf("config: job[%d] %q missing command", i, j.Name)
		}
	}
	return nil
}
