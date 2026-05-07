// Package webhook provides HTTP webhook notifications for cron job failures.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Config holds webhook notifier configuration.
type Config struct {
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Timeout time.Duration     `yaml:"timeout"`
}

// Payload is the JSON body sent on job failure.
type Payload struct {
	JobName  string    `json:"job_name"`
	Command  string    `json:"command"`
	ExitCode int       `json:"exit_code"`
	Output   string    `json:"output,omitempty"`
	FailedAt time.Time `json:"failed_at"`
}

// Notifier sends webhook notifications.
type Notifier struct {
	cfg    Config
	client *http.Client
}

// New creates a new webhook Notifier with the given config.
func New(cfg Config) *Notifier {
	if cfg.Method == "" {
		cfg.Method = http.MethodPost
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Notifier{
		cfg:    cfg,
		client: &http.Client{Timeout: timeout},
	}
}

// Send dispatches a webhook for the given payload.
func (n *Notifier) Send(ctx context.Context, p Payload) error {
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, n.cfg.Method, n.cfg.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range n.cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d from %s", resp.StatusCode, n.cfg.URL)
	}
	return nil
}
