package scheduler

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

type countingAlerter struct {
	calls atomic.Int32
	err   error
}

func (c *countingAlerter) Alert(_ Result) error {
	c.calls.Add(1)
	return c.err
}

func failResult(name string) Result {
	return Result{
		Job:    config.Job{Name: name, Command: "false"},
		Output: runner.Result{ExitCode: 1},
	}
}

func TestThrottle_FirstAlertIsForwarded(t *testing.T) {
	inner := &countingAlerter{}
	th := NewThrottledAlerter(inner, 5*time.Minute)

	if err := th.Alert(failResult("job1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestThrottle_SecondAlertSuppressedWithinCooldown(t *testing.T) {
	inner := &countingAlerter{}
	th := NewThrottledAlerter(inner, 5*time.Minute)

	_ = th.Alert(failResult("job1"))
	_ = th.Alert(failResult("job1"))

	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call after suppression, got %d", got)
	}
}

func TestThrottle_AlertForwardedAfterCooldown(t *testing.T) {
	inner := &countingAlerter{}
	th := NewThrottledAlerter(inner, 50*time.Millisecond)

	_ = th.Alert(failResult("job1"))
	time.Sleep(80 * time.Millisecond)
	_ = th.Alert(failResult("job1"))

	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls after cooldown, got %d", got)
	}
}

func TestThrottle_DifferentJobsAreIndependent(t *testing.T) {
	inner := &countingAlerter{}
	th := NewThrottledAlerter(inner, 5*time.Minute)

	_ = th.Alert(failResult("job1"))
	_ = th.Alert(failResult("job2"))

	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls for distinct jobs, got %d", got)
	}
}

func TestThrottle_ResetAllowsImmediateResend(t *testing.T) {
	inner := &countingAlerter{}
	th := NewThrottledAlerter(inner, 5*time.Minute)

	_ = th.Alert(failResult("job1"))
	th.Reset("job1")
	_ = th.Alert(failResult("job1"))

	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls after reset, got %d", got)
	}
}

func TestThrottle_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("smtp down")
	inner := &countingAlerter{err: sentinel}
	th := NewThrottledAlerter(inner, 5*time.Minute)

	err := th.Alert(failResult("job1"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
