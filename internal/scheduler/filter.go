package scheduler

import (
	"context"
	"strings"

	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

// FilterFunc decides whether a result should be forwarded to the inner alerter.
type FilterFunc func(result runner.Result) bool

// FilteredAlerter wraps an Alerter and only forwards results that pass the filter.
type FilteredAlerter struct {
	inner  Alerter
	filter FilterFunc
}

// NewFilteredAlerter returns an Alerter that calls inner only when filter returns true.
func NewFilteredAlerter(inner Alerter, filter FilterFunc) *FilteredAlerter {
	return &FilteredAlerter{inner: inner, filter: filter}
}

// Alert forwards the result to the inner alerter only when the filter passes.
func (f *FilteredAlerter) Alert(ctx context.Context, result runner.Result) error {
	if !f.filter(result) {
		return nil
	}
	return f.inner.Alert(ctx, result)
}

// ExitCodeFilter returns a FilterFunc that allows only results whose exit code
// is in the provided set. An empty set allows all exit codes.
func ExitCodeFilter(codes []int) FilterFunc {
	if len(codes) == 0 {
		return func(_ runner.Result) bool { return true }
	}
	set := make(map[int]struct{}, len(codes))
	for _, c := range codes {
		set[c] = struct{}{}
	}
	return func(r runner.Result) bool {
		_, ok := set[r.ExitCode]
		return ok
	}
}

// OutputContainsFilter returns a FilterFunc that allows results whose combined
// output contains any of the provided substrings (case-insensitive).
func OutputContainsFilter(patterns []string) FilterFunc {
	if len(patterns) == 0 {
		return func(_ runner.Result) bool { return true }
	}
	lower := make([]string, len(patterns))
	for i, p := range patterns {
		lower[i] = strings.ToLower(p)
	}
	return func(r runner.Result) bool {
		haystack := strings.ToLower(r.Output)
		for _, p := range lower {
			if strings.Contains(haystack, p) {
				return true
			}
		}
		return false
	}
}

// FilterFromJob builds a FilteredAlerter from job-level config when alert
// filters are configured. If no filters are set the original alerter is returned
// unchanged.
func FilterFromJob(job config.Job, inner Alerter) Alerter {
	hasExitCodes := len(job.AlertFilter.ExitCodes) > 0
	hasPatterns := len(job.AlertFilter.OutputContains) > 0

	if !hasExitCodes && !hasPatterns {
		return inner
	}

	var filters []FilterFunc
	if hasExitCodes {
		filters = append(filters, ExitCodeFilter(job.AlertFilter.ExitCodes))
	}
	if hasPatterns {
		filters = append(filters, OutputContainsFilter(job.AlertFilter.OutputContains))
	}

	combined := func(r runner.Result) bool {
		for _, f := range filters {
			if f(r) {
				return true
			}
		}
		return false
	}
	return NewFilteredAlerter(inner, combined)
}
