package history

import (
	"testing"
	"time"
)

func seedEntries(t *testing.T, s *Store, jobName string, entries []Entry) {
	t.Helper()
	for _, e := range entries {
		if err := s.Record(e); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
}

func TestSummarize_BasicCounts(t *testing.T) {
	s := tempStore(t)
	now := time.Now()

	seedEntries(t, s, "backup", []Entry{
		{JobName: "backup", StartedAt: now.Add(-1 * time.Hour), Duration: 2 * time.Second, Success: true},
		{JobName: "backup", StartedAt: now.Add(-2 * time.Hour), Duration: 3 * time.Second, Success: false},
		{JobName: "backup", StartedAt: now.Add(-3 * time.Hour), Duration: 1 * time.Second, Success: true},
	})

	sm, err := s.Summarize("backup", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.Total != 3 {
		t.Errorf("Total: want 3, got %d", sm.Total)
	}
	if sm.Successes != 2 {
		t.Errorf("Successes: want 2, got %d", sm.Successes)
	}
	if sm.Failures != 1 {
		t.Errorf("Failures: want 1, got %d", sm.Failures)
	}
	if sm.SuccessRate < 66.6 || sm.SuccessRate > 66.8 {
		t.Errorf("SuccessRate: want ~66.7, got %.2f", sm.SuccessRate)
	}
}

func TestSummarize_WindowFiltersOldEntries(t *testing.T) {
	s := tempStore(t)
	now := time.Now()

	seedEntries(t, s, "cleanup", []Entry{
		{JobName: "cleanup", StartedAt: now.Add(-30 * time.Minute), Duration: time.Second, Success: true},
		{JobName: "cleanup", StartedAt: now.Add(-25 * time.Hour), Duration: time.Second, Success: false},
	})

	sm, err := s.Summarize("cleanup", 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.Total != 1 {
		t.Errorf("Total: want 1, got %d", sm.Total)
	}
	if sm.Successes != 1 {
		t.Errorf("Successes: want 1, got %d", sm.Successes)
	}
}

func TestSummarize_UnknownJobReturnsEmpty(t *testing.T) {
	s := tempStore(t)
	sm, err := s.Summarize("ghost", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.Total != 0 {
		t.Errorf("expected zero total for unknown job")
	}
}

func TestSummarizeAll_CoversAllJobs(t *testing.T) {
	s := tempStore(t)
	now := time.Now()

	seedEntries(t, s, "alpha", []Entry{
		{JobName: "alpha", StartedAt: now.Add(-1 * time.Hour), Duration: time.Second, Success: true},
	})
	seedEntries(t, s, "beta", []Entry{
		{JobName: "beta", StartedAt: now.Add(-2 * time.Hour), Duration: time.Second, Success: false},
	})

	summaries, err := s.SummarizeAll(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 2 {
		t.Errorf("want 2 summaries, got %d", len(summaries))
	}
}
