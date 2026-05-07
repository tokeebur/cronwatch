package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/history"
)

func TestHandleMetrics_MethodNotAllowed(t *testing.T) {
	store := makeStore(t)
	srv := New(store, DefaultConfig())

	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleMetrics_ReturnsPrometheusLines(t *testing.T) {
	store := makeStore(t)

	now := time.Now()
	for i := 0; i < 3; i++ {
		store.Record(history.Entry{
			JobName:    "backup",
			StartedAt:  now.Add(-time.Duration(i) * time.Minute),
			DurationMs: 120,
			Success:    i != 0, // first entry is a failure
		})
	}

	srv := New(store, DefaultConfig())
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `cronwatch_runs_total{job="backup"}`) {
		t.Errorf("missing runs_total line; body: %s", body)
	}
	if !strings.Contains(body, `cronwatch_failures_total{job="backup"}`) {
		t.Errorf("missing failures_total line; body: %s", body)
	}
	if !strings.Contains(body, fmt.Sprintf("%d", 3)) {
		t.Errorf("expected total count 3 in body; body: %s", body)
	}
}

func TestHandleMetrics_ContentType(t *testing.T) {
	store := makeStore(t)
	srv := New(store, DefaultConfig())

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected text/plain content-type, got %q", ct)
	}
}
