package scheduler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/runner"
	"github.com/cronwatch/cronwatch/internal/scheduler"
	"github.com/cronwatch/cronwatch/internal/webhook"
)

func TestWebhookAlerter_SendsOnAlert(t *testing.T) {
	var received webhook.Payload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	alerter := scheduler.NewWebhookAlerter(webhook.Config{URL: ts.URL})
	alerter.Alert(runner.Result{
		JobName:  "nightly-sync",
		Command:  "/usr/local/bin/sync.sh",
		ExitCode: 2,
		Output:   "connection refused",
	})

	// Allow the HTTP call to complete.
	time.Sleep(50 * time.Millisecond)

	if received.JobName != "nightly-sync" {
		t.Errorf("got job_name %q, want %q", received.JobName, "nightly-sync")
	}
	if received.ExitCode != 2 {
		t.Errorf("got exit_code %d, want 2", received.ExitCode)
	}
	if received.Output != "connection refused" {
		t.Errorf("got output %q, want %q", received.Output, "connection refused")
	}
}

func TestWebhookAlerter_DoesNotPanicOnBadURL(t *testing.T) {
	alerter := scheduler.NewWebhookAlerter(webhook.Config{URL: "://invalid"})
	// Should log the error but not panic.
	alerter.Alert(runner.Result{JobName: "test", ExitCode: 1})
}
