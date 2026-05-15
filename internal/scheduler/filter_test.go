package scheduler

import (
	"context"
	"errors"
	"testing"

	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

type captureAlerter struct {
	calls []runner.Result
	err   error
}

func (c *captureAlerter) Alert(_ context.Context, r runner.Result) error {
	c.calls = append(c.calls, r)
	return c.err
}

func filterResult(exit int, output string) runner.Result {
	return runner.Result{JobName: "test", ExitCode: exit, Output: output}
}

func TestFilteredAlerter_PassesWhenFilterTrue(t *testing.T) {
	cap := &captureAlerter{}
	a := NewFilteredAlerter(cap, func(_ runner.Result) bool { return true })
	_ = a.Alert(context.Background(), filterResult(1, ""))
	if len(cap.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(cap.calls))
	}
}

func TestFilteredAlerter_SuppressesWhenFilterFalse(t *testing.T) {
	cap := &captureAlerter{}
	a := NewFilteredAlerter(cap, func(_ runner.Result) bool { return false })
	_ = a.Alert(context.Background(), filterResult(1, ""))
	if len(cap.calls) != 0 {
		t.Fatalf("expected 0 calls, got %d", len(cap.calls))
	}
}

func TestExitCodeFilter_MatchesCode(t *testing.T) {
	f := ExitCodeFilter([]int{1, 2})
	if !f(filterResult(1, "")) {
		t.Error("expected exit code 1 to match")
	}
	if f(filterResult(3, "")) {
		t.Error("expected exit code 3 not to match")
	}
}

func TestExitCodeFilter_EmptyAllowsAll(t *testing.T) {
	f := ExitCodeFilter(nil)
	if !f(filterResult(99, "")) {
		t.Error("empty filter should allow all exit codes")
	}
}

func TestOutputContainsFilter_MatchesCaseInsensitive(t *testing.T) {
	f := OutputContainsFilter([]string{"ERROR"})
	if !f(filterResult(0, "some error occurred")) {
		t.Error("expected case-insensitive match")
	}
	if f(filterResult(0, "all good")) {
		t.Error("expected no match")
	}
}

func TestOutputContainsFilter_EmptyAllowsAll(t *testing.T) {
	f := OutputContainsFilter(nil)
	if !f(filterResult(0, "")) {
		t.Error("empty pattern list should allow all results")
	}
}

func TestFilterFromJob_NoFiltersPassthrough(t *testing.T) {
	cap := &captureAlerter{}
	job := config.Job{}
	a := FilterFromJob(job, cap)
	if a != cap {
		t.Error("expected original alerter when no filters configured")
	}
}

func TestFilterFromJob_ExitCodeFilterApplied(t *testing.T) {
	cap := &captureAlerter{}
	job := config.Job{AlertFilter: config.AlertFilter{ExitCodes: []int{2}}}
	a := FilterFromJob(job, cap)
	_ = a.Alert(context.Background(), filterResult(1, ""))
	if len(cap.calls) != 0 {
		t.Error("exit code 1 should be suppressed")
	}
	_ = a.Alert(context.Background(), filterResult(2, ""))
	if len(cap.calls) != 1 {
		t.Error("exit code 2 should be forwarded")
	}
}

func TestFilteredAlerter_PropagatesError(t *testing.T) {
	sentinel := errors.New("alert failed")
	cap := &captureAlerter{err: sentinel}
	a := NewFilteredAlerter(cap, func(_ runner.Result) bool { return true })
	err := a.Alert(context.Background(), filterResult(1, ""))
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
