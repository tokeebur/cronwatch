package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/api"
	"github.com/user/cronwatch/internal/history"
)

func makeStore(t *testing.T) *history.Store {
	t.Helper()
	dir := t.TempDir()
	store, err := history.New(dir)
	if err != nil {
		t.Fatalf("history.New: %v", err)
	}
	return store
}

func TestHandleHealth(t *testing.T) {
	s := api.New(":0", makeStore(t))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %q", body["status"])
	}
}

func TestHandleHistory_MissingParam(t *testing.T) {
	s := api.New(":0", makeStore(t))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleHistory_UnknownJob(t *testing.T) {
	s := api.New(":0", makeStore(t))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history?job=ghost", nil)
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleStatus_ReturnsSummaries(t *testing.T) {
	store := makeStore(t)
	store.Record("backup", history.Entry{Success: true, StartedAt: time.Now(), Duration: time.Second})
	s := api.New(":0", store)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var summaries []history.Summary
	if err := json.NewDecoder(rec.Body).Decode(&summaries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(summaries) == 0 {
		t.Error("expected at least one summary")
	}
}
