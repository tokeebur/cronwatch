package webhook_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/webhook"
)

func TestSend_Success(t *testing.T) {
	var received webhook.Payload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := webhook.New(webhook.Config{URL: ts.URL})
	p := webhook.Payload{
		JobName:  "backup",
		Command:  "/usr/bin/backup.sh",
		ExitCode: 1,
		Output:   "disk full",
		FailedAt: time.Now(),
	}
	if err := n.Send(context.Background(), p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.JobName != "backup" {
		t.Errorf("got job_name %q, want %q", received.JobName, "backup")
	}
	if received.ExitCode != 1 {
		t.Errorf("got exit_code %d, want 1", received.ExitCode)
	}
}

func TestSend_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := webhook.New(webhook.Config{URL: ts.URL})
	err := n.Send(context.Background(), webhook.Payload{JobName: "test"})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestSend_CustomHeaders(t *testing.T) {
	var gotAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	n := webhook.New(webhook.Config{
		URL:     ts.URL,
		Headers: map[string]string{"Authorization": "Bearer secret"},
	})
	if err := n.Send(context.Background(), webhook.Payload{JobName: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "Bearer secret" {
		t.Errorf("got Authorization %q, want %q", gotAuth, "Bearer secret")
	}
}

func TestSend_InvalidURL(t *testing.T) {
	n := webhook.New(webhook.Config{URL: "://bad-url"})
	err := n.Send(context.Background(), webhook.Payload{JobName: "x"})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestSend_DefaultMethod(t *testing.T) {
	var gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := webhook.New(webhook.Config{URL: ts.URL})
	_ = n.Send(context.Background(), webhook.Payload{})
	if gotMethod != http.MethodPost {
		t.Errorf("got method %q, want POST", gotMethod)
	}
}
