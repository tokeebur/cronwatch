package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/runner"
)

func makeJob(name string) runner.Job {
	return runner.Job{Name: name, Command: "echo hi"}
}

type countingRunner struct{ calls int }

func (c *countingRunner) Run(_ context.Context, job runner.Job) runner.Result {
	c.calls++
	return runner.Result{Job: job}
}

func TestRateLimit_FirstRunAlwaysExecutes(t *testing.T) {
	inner := &countingRunner{}
	rl := NewRateLimitedRunner(inner, 10*time.Second)

	res := rl.Run(context.Background(), makeJob("myjob"))

	if res.Skipped {
		t.Fatal("expected first run to execute, got skipped")
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", inner.calls)
	}
}

func TestRateLimit_SecondRunSkippedWithinInterval(t *testing.T) {
	inner := &countingRunner{}
	rl := NewRateLimitedRunner(inner, 10*time.Second)

	rl.Run(context.Background(), makeJob("myjob"))
	res := rl.Run(context.Background(), makeJob("myjob"))

	if !res.Skipped {
		t.Fatal("expected second run to be skipped within interval")
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", inner.calls)
	}
}

func TestRateLimit_RunsAfterIntervalExpires(t *testing.T) {
	inner := &countingRunner{}
	rl := NewRateLimitedRunner(inner, 50*time.Millisecond)

	rl.Run(context.Background(), makeJob("myjob"))
	time.Sleep(60 * time.Millisecond)
	res := rl.Run(context.Background(), makeJob("myjob"))

	if res.Skipped {
		t.Fatal("expected run to proceed after interval expired")
	}
	if inner.calls != 2 {
		t.Fatalf("expected 2 inner calls, got %d", inner.calls)
	}
}

func TestRateLimit_ZeroIntervalNeverSkips(t *testing.T) {
	inner := &countingRunner{}
	rl := NewRateLimitedRunner(inner, 0)

	for i := 0; i < 5; i++ {
		res := rl.Run(context.Background(), makeJob("myjob"))
		if res.Skipped {
			t.Fatalf("run %d skipped with zero interval", i+1)
		}
	}
	if inner.calls != 5 {
		t.Fatalf("expected 5 inner calls, got %d", inner.calls)
	}
}

func TestRateLimitFromJob_ParsesDuration(t *testing.T) {
	job := runner.Job{Name: "j", Meta: map[string]string{"rate_limit_interval": "2m"}}
	d, err := RateLimitFromJob(job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 2*time.Minute {
		t.Fatalf("expected 2m, got %s", d)
	}
}

func TestRateLimitFromJob_MissingKey(t *testing.T) {
	job := runner.Job{Name: "j", Meta: map[string]string{}}
	d, err := RateLimitFromJob(job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 0 {
		t.Fatalf("expected 0, got %s", d)
	}
}

func TestRateLimitFromJob_InvalidDuration(t *testing.T) {
	job := runner.Job{Name: "j", Meta: map[string]string{"rate_limit_interval": "not-a-duration"}}
	_, err := RateLimitFromJob(job)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}
