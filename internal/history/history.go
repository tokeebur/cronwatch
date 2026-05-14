// Package history records cron job execution results to a local file-backed store.
package history

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry holds the result of a single job execution.
type Entry struct {
	JobName   string        `json:"job_name"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration_ns"`
	ExitCode  int           `json:"exit_code"`
	Output    string        `json:"output,omitempty"`
	Error     string        `json:"error,omitempty"`
}

// Store persists job execution history.
type Store struct {
	mu      sync.Mutex
	path    string
	entries []Entry
}

// New opens (or creates) a history store at the given file path.
func New(path string) (*Store, error) {
	s := &Store{path: path}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &s.entries); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Record appends an entry and flushes to disk.
func (s *Store) Record(e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, e)
	return s.flush()
}

// Last returns the most recent entry for the given job, or false if none.
func (s *Store) Last(jobName string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.entries) - 1; i >= 0; i-- {
		if s.entries[i].JobName == jobName {
			return s.entries[i], true
		}
	}
	return Entry{}, false
}

// All returns a copy of all recorded entries.
func (s *Store) All() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Entry, len(s.entries))
	copy(out, s.entries)
	return out
}

// ForJob returns a copy of all recorded entries for the given job name,
// in chronological order.
func (s *Store) ForJob(jobName string) []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Entry
	for _, e := range s.entries {
		if e.JobName == jobName {
			out = append(out, e)
		}
	}
	return out
}

func (s *Store) flush() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}
