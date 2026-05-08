package scheduler

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/history"
	"github.com/cronwatch/cronwatch/internal/runner"
)

func tempHistory(t *testing.T) *history.Store {
	t.Helper()
	f, err := os.CreateTemp("", "cronwatch-hist-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	store, err := history.New(f.Name())
	if err != nil {
		t.Fatalf("history.New: %v", err)
	}
	return store
}

func TestHistoryRecorder_RecordsSuccess(t *testing.T) {
	store := tempHistory(t)
	rec := NewHistoryRecorder(store)

	result := runner.Result{
		JobName:   "backup",
		StartedAt: time.Now(),
		Duration:  2 * time.Second,
		ExitCode:  0,
		Output:    "done",
		Err:       nil,
	}
	rec.Record(result)

	entry, err := store.Last("backup")
	if err != nil {
		t.Fatalf("Last: %v", err)
	}
	if !entry.Success {
		t.Error("expected entry to be marked successful")
	}
	if entry.JobName != "backup" {
		t.Errorf("got job name %q, want %q", entry.JobName, "backup")
	}
	if entry.Error != "" {
		t.Errorf("expected empty error string, got %q", entry.Error)
	}
}

func TestHistoryRecorder_RecordsFailure(t *testing.T) {
	store := tempHistory(t)
	rec := NewHistoryRecorder(store)

	result := runner.Result{
		JobName:   "cleanup",
		StartedAt: time.Now(),
		Duration:  500 * time.Millisecond,
		ExitCode:  1,
		Output:    "",
		Err:       errors.New("exit status 1"),
	}
	rec.Record(result)

	entry, err := store.Last("cleanup")
	if err != nil {
		t.Fatalf("Last: %v", err)
	}
	if entry.Success {
		t.Error("expected entry to be marked as failure")
	}
	if entry.Error != "exit status 1" {
		t.Errorf("got error %q, want %q", entry.Error, "exit status 1")
	}
	if entry.ExitCode != 1 {
		t.Errorf("got exit code %d, want 1", entry.ExitCode)
	}
}
