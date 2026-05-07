package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/history"
)

func tempStore(t *testing.T) (*history.Store, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	s, err := history.New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s, path
}

func TestRecord_And_Last(t *testing.T) {
	s, _ := tempStore(t)
	e := history.Entry{
		JobName:   "backup",
		StartedAt: time.Now(),
		Duration:  2 * time.Second,
		ExitCode:  0,
	}
	if err := s.Record(e); err != nil {
		t.Fatalf("Record: %v", err)
	}
	got, ok := s.Last("backup")
	if !ok {
		t.Fatal("expected entry, got none")
	}
	if got.ExitCode != 0 {
		t.Errorf("exit code: want 0, got %d", got.ExitCode)
	}
}

func TestLast_UnknownJob(t *testing.T) {
	s, _ := tempStore(t)
	_, ok := s.Last("nonexistent")
	if ok {
		t.Error("expected no entry for unknown job")
	}
}

func TestPersistence(t *testing.T) {
	s, path := tempStore(t)
	e := history.Entry{JobName: "cleanup", ExitCode: 1, Error: "timeout"}
	if err := s.Record(e); err != nil {
		t.Fatalf("Record: %v", err)
	}

	// Re-open the store from the same file.
	s2, err := history.New(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	got, ok := s2.Last("cleanup")
	if !ok {
		t.Fatal("entry not persisted")
	}
	if got.Error != "timeout" {
		t.Errorf("error field: want %q, got %q", "timeout", got.Error)
	}
}

func TestAll_ReturnsAllEntries(t *testing.T) {
	s, _ := tempStore(t)
	for _, name := range []string{"job-a", "job-b", "job-a"} {
		_ = s.Record(history.Entry{JobName: name})
	}
	if len(s.All()) != 3 {
		t.Errorf("want 3 entries, got %d", len(s.All()))
	}
}

func TestNew_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(path, []byte("not-json{"), 0o644)
	_, err := history.New(path)
	if err == nil {
		t.Error("expected error for corrupt JSON, got nil")
	}
}
