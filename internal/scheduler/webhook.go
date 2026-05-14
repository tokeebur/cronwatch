package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/cronwatch/cronwatch/internal/runner"
	"github.com/cronwatch/cronwatch/internal/webhook"
)

// WebhookAlerter wraps a webhook.Notifier to satisfy the alerter interface
// used by the Scheduler.
type WebhookAlerter struct {
	notifier *webhook.Notifier
	timeout  time.Duration
}

// NewWebhookAlerter creates a WebhookAlerter from a webhook.Config.
func NewWebhookAlerter(cfg webhook.Config) *WebhookAlerter {
	return &WebhookAlerter{
		notifier: webhook.New(cfg),
		timeout:  10 * time.Second,
	}
}

// Alert sends a webhook notification for a failed job result.
func (w *WebhookAlerter) Alert(result runner.Result) {
	p := webhook.Payload{
		JobName:  result.JobName,
		Command:  result.Command,
		ExitCode: result.ExitCode,
		Output:   result.Output,
		FailedAt: time.Now(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), w.timeout)
	defer cancel()
	if err := w.notifier.Send(ctx, p); err != nil {
		log.Printf("webhook alert failed for job %q: %v", result.JobName, err)
	}
}
