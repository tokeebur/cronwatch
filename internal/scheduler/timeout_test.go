package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

// slowRunner simulates a job that sleeps until its context is cancelled.
type slowRunner struct {
	sleep time.Duration
}

func (s *slowRunner) Run(ctx context.Context, job config.Job) runner.Result {
	select {
	case <-time.After(s.sleep):
		return runner.Result{Job: job}
	case <-ctx.Done():
		return runner.Result{Job: job, Error: ctx.Err(), ExitCode: -1}
	}
}

func TestTimeoutRunner_CompletesWithinDeadline(t *testing.T) {
	inner := &slowRunner{sleep: 10 * time.Millisecond}
	tr := NewTimeoutRunner(inner, 500*time.Millisecond)

	job := config.Job{Name: "fast-job", Command: "echo hi"}
	res := tr.Run(context.Background(), job)

	if res.Error != nil {
		t.Fatalf("expected no error, got %v", res.Error)
	}
}

func TestTimeoutRunner_CancelsSlowJob(t *testing.T) {
	inner := &slowRunner{sleep: 2 * time.Second}
	tr := NewTimeoutRunner(inner, 50*time.Millisecond)

	job := config.Job{Name: "slow-job", Command: "sleep 2"}
	res := tr.Run(context.Background(), job)

	if res.Error == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if res.ExitCode != -1 {
		t.Errorf("expected exit code -1, got %d", res.ExitCode)
	}
}

func TestTimeoutRunner_ZeroTimeoutIsNoop(t *testing.T) {
	inner := &slowRunner{sleep: 10 * time.Millisecond}
	wrapped := NewTimeoutRunner(inner, 0)

	if _, ok := wrapped.(*TimeoutRunner); ok {
		t.Error("expected zero timeout to return inner runner unwrapped")
	}
}

func TestTimeoutFromJob_ParsesDuration(t *testing.T) {
	inner := &slowRunner{sleep: 2 * time.Second}
	job := config.Job{Name: "j", Command: "sleep 2", Timeout: "50ms"}

	wrapped := TimeoutFromJob(inner, job)
	res := wrapped.Run(context.Background(), job)

	if res.Error == nil {
		t.Fatal("expected timeout error from job config")
	}
}

func TestTimeoutFromJob_EmptyTimeoutIsNoop(t *testing.T) {
	inner := &slowRunner{sleep: 10 * time.Millisecond}
	job := config.Job{Name: "j", Command: "echo", Timeout: ""}

	wrapped := TimeoutFromJob(inner, job)
	if _, ok := wrapped.(*TimeoutRunner); ok {
		t.Error("expected empty timeout to return inner runner unwrapped")
	}
}

func TestTimeoutFromJob_InvalidDurationIsNoop(t *testing.T) {
	inner := &slowRunner{sleep: 10 * time.Millisecond}
	job := config.Job{Name: "j", Command: "echo", Timeout: "not-a-duration"}

	wrapped := TimeoutFromJob(inner, job)
	if _, ok := wrapped.(*TimeoutRunner); ok {
		t.Error("expected invalid duration to return inner runner unwrapped")
	}
}
