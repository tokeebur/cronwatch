package history

import "time"

// JobSummary holds aggregated statistics for a single job over a time window.
type JobSummary struct {
	JobName      string
	Total        int
	Successes    int
	Failures     int
	LastRun      time.Time
	LastDuration time.Duration
	SuccessRate  float64
}

// Summarize returns a JobSummary for the named job, considering only entries
// that occurred within the given duration window (e.g. 24*time.Hour).
// If window is zero, all entries are included.
func (s *Store) Summarize(jobName string, window time.Duration) (JobSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, ok := s.data[jobName]
	if !ok {
		return JobSummary{JobName: jobName}, nil
	}

	cutoff := time.Time{}
	if window > 0 {
		cutoff = time.Now().Add(-window)
	}

	var summary JobSummary
	summary.JobName = jobName

	for _, e := range entries {
		if !cutoff.IsZero() && e.StartedAt.Before(cutoff) {
			continue
		}
		summary.Total++
		if e.Success {
			summary.Successes++
		} else {
			summary.Failures++
		}
		if e.StartedAt.After(summary.LastRun) {
			summary.LastRun = e.StartedAt
			summary.LastDuration = e.Duration
		}
	}

	if summary.Total > 0 {
		summary.SuccessRate = float64(summary.Successes) / float64(summary.Total) * 100
	}

	return summary, nil
}

// SummarizeAll returns a JobSummary for every job known to the store.
func (s *Store) SummarizeAll(window time.Duration) ([]JobSummary, error) {
	s.mu.RLock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	s.mu.RUnlock()

	summaries := make([]JobSummary, 0, len(keys))
	for _, k := range keys {
		sm, err := s.Summarize(k, window)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, sm)
	}
	return summaries, nil
}
