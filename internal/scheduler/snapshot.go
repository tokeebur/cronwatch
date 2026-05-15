package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/user/cronwatch/internal/runner"
)

// SnapshotRunner wraps a Runner and retains the most recent Result in memory.
// It is safe for concurrent access and is useful for exposing live job state
// via the API without hitting persistent storage.
type SnapshotRunner struct {
	inner  runner.Runner
	name   string
	mu     sync.RWMutex
	latest *runner.Result
}

// NewSnapshotRunner returns a SnapshotRunner that delegates to inner and
// stores each result so callers can inspect it later.
func NewSnapshotRunner(name string, inner runner.Runner) *SnapshotRunner {
	return &SnapshotRunner{name: name, inner: inner}
}

// Run executes the inner runner and captures the result.
func (s *SnapshotRunner) Run(ctx context.Context) runner.Result {
	res := s.inner.Run(ctx)
	s.mu.Lock()
	s.latest = &res
	s.mu.Unlock()
	return res
}

// Latest returns the most recent Result and true, or a zero Result and false
// if the job has never been run.
func (s *SnapshotRunner) Latest() (runner.Result, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.latest == nil {
		return runner.Result{}, false
	}
	return *s.latest, true
}

// SnapshotStore holds SnapshotRunners indexed by job name and provides
// aggregate access for the API layer.
type SnapshotStore struct {
	mu      sync.RWMutex
	runners map[string]*SnapshotRunner
}

// NewSnapshotStore returns an empty SnapshotStore.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{runners: make(map[string]*SnapshotRunner)}
}

// Register adds a SnapshotRunner to the store.
func (ss *SnapshotStore) Register(sr *SnapshotRunner) {
	ss.mu.Lock()
	ss.runners[sr.name] = sr
	ss.mu.Unlock()
}

// Snapshot holds the last-known state of a single job.
type Snapshot struct {
	Job     string
	LastRun time.Time
	Success bool
	Output  string
}

// All returns a Snapshot for every registered job that has run at least once.
func (ss *SnapshotStore) All() []Snapshot {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	out := make([]Snapshot, 0, len(ss.runners))
	for name, sr := range ss.runners {
		res, ok := sr.Latest()
		if !ok {
			continue
		}
		out = append(out, Snapshot{
			Job:     name,
			LastRun: res.FinishedAt,
			Success: res.ExitCode == 0,
			Output:  res.Output,
		})
	}
	return out
}
