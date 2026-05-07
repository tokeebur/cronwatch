package history

import (
	"testing"
	"time"
)

func TestPrune_RemovesOldEntries(t *testing.T) {
	dir := tempStore(t)
	s := New(dir)

	old := Entry{
		JobName:    "old-job",
		Success:    true,
		StartedAt:  time.Now().Add(-48 * time.Hour),
		FinishedAt: time.Now().Add(-47 * time.Hour),
	}
	recent := Entry{
		JobName:    "recent-job",
		Success:    true,
		StartedAt:  time.Now().Add(-30 * time.Minute),
		FinishedAt: time.Now().Add(-29 * time.Minute),
	}

	if err := s.Record(old); err != nil {
		t.Fatalf("record old: %v", err)
	}
	if err := s.Record(recent); err != nil {
		t.Fatalf("record recent: %v", err)
	}

	cleaner := NewCleaner(s, 24*time.Hour)
	n, err := cleaner.Prune()
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 removed, got %d", n)
	}

	if _, err := s.Last("old-job"); err == nil {
		t.Error("expected old-job to be pruned, but it still exists")
	}
	if _, err := s.Last("recent-job"); err != nil {
		t.Errorf("expected recent-job to remain: %v", err)
	}
}

func TestPrune_KeepsAllWhenNoneExpired(t *testing.T) {
	dir := tempStore(t)
	s := New(dir)

	for i := 0; i < 3; i++ {
		e := Entry{
			JobName:    "job",
			Success:    true,
			StartedAt:  time.Now().Add(-time.Minute),
			FinishedAt: time.Now(),
		}
		if err := s.Record(e); err != nil {
			t.Fatalf("record: %v", err)
		}
	}

	cleaner := NewCleaner(s, 24*time.Hour)
	n, err := cleaner.Prune()
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 removed, got %d", n)
	}

	all, err := s.All("job")
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 entries, got %d", len(all))
	}
}

func TestPrune_EmptyStore(t *testing.T) {
	dir := tempStore(t)
	s := New(dir)

	cleaner := NewCleaner(s, 24*time.Hour)
	n, err := cleaner.Prune()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 removed, got %d", n)
	}
}
