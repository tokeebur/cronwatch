package api

import "errors"

// Config holds configuration for the HTTP API server.
type Config struct {
	// Enabled controls whether the API server starts at all.
	Enabled bool `yaml:"enabled"`
	// Addr is the TCP address to listen on, e.g. ":8080".
	Addr string `yaml:"addr"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Enabled: false,
		Addr:    ":8080",
	}
}

// Validate checks that the Config is usable.
func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Addr == "" {
		return errors.New("api.addr must not be empty when api is enabled")
	}
	return nil
}
