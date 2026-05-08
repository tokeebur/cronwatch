package scheduler

import (
	"time"

	"github.com/cronwatch/cronwatch/internal/history"
	"github.com/cronwatch/cronwatch/internal/runner"
)

// HistoryRecorder records job run results into a history store.
type HistoryRecorder struct {
	store *history.Store
}

// NewHistoryRecorder creates a HistoryRecorder backed by the given store.
func NewHistoryRecorder(store *history.Store) *HistoryRecorder {
	return &HistoryRecorder{store: store}
}

// Record persists the result of a job run to the history store.
func (h *HistoryRecorder) Record(result runner.Result) {
	entry := history.Entry{
		JobName:   result.JobName,
		StartedAt: result.StartedAt,
		Duration:  result.Duration,
		ExitCode:  result.ExitCode,
		Output:    result.Output,
		Error:     errorString(result.Err),
		Success:   result.Err == nil && result.ExitCode == 0,
		RecordedAt: time.Now(),
	}
	_ = h.store.Record(entry)
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
