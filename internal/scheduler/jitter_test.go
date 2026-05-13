package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/cronwatch/internal/config"
)

// fixedRunner is a test double that returns a pre-set Result immediately.
type fixedRunner struct {
	result Result
	called int
}

func (f *fixedRunner) Run(_ context.Context, job config.Job) Result {
	f.called++
	f.result.Job = job
	return f.result
}

func TestJitterRunner_DelegatesToInner(t *testing.T) {
	inner := &fixedRunner{result: Result{ExitCode: 0}}
	jr := NewJitterRunner(inner, 0) // zero jitter — no sleep

	job := config.Job{Name: "noop", Command: "true"}
	res := jr.Run(context.Background(), job)

	if inner.called != 1 {
		t.Fatalf("expected inner called once, got %d", inner.called)
	}
	if res.Job.Name != "noop" {
		t.Errorf("unexpected job name: %s", res.Job.Name)
	}
}

func TestJitterRunner_ZeroJitterNoDelay(t *testing.T) {
	inner := &fixedRunner{result: Result{ExitCode: 0}}
	jr := NewJitterRunner(inner, 0)

	start := time.Now()
	jr.Run(context.Background(), config.Job{Name: "fast"})
	elapsed := time.Since(start)

	if elapsed > 50*time.Millisecond {
		t.Errorf("zero jitter run took too long: %v", elapsed)
	}
}

func TestJitterRunner_ContextCancelledDuringJitter(t *testing.T) {
	inner := &fixedRunner{result: Result{ExitCode: 0}}
	// Use a large jitter so the sleep won't finish before cancellation.
	jr := NewJitterRunner(inner, 10*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	res := jr.Run(ctx, config.Job{Name: "cancelled"})

	if inner.called != 0 {
		t.Error("inner runner should not have been called after context cancellation")
	}
	if res.Error == nil {
		t.Error("expected a non-nil error from cancelled context")
	}
}

func TestJitterFromJob_ParsesDuration(t *testing.T) {
	job := config.Job{Jitter: "500ms"}
	d := JitterFromJob(job)
	if d != 500*time.Millisecond {
		t.Errorf("expected 500ms, got %v", d)
	}
}

func TestJitterFromJob_EmptyReturnsZero(t *testing.T) {
	if d := JitterFromJob(config.Job{}); d != 0 {
		t.Errorf("expected 0, got %v", d)
	}
}

func TestJitterFromJob_InvalidReturnsZero(t *testing.T) {
	if d := JitterFromJob(config.Job{Jitter: "not-a-duration"}); d != 0 {
		t.Errorf("expected 0, got %v", d)
	}
}
