package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/nwlnexus/cronwatch/internal/config"
	"github.com/nwlnexus/cronwatch/internal/runner"
)

func cooldownJob(cooldown string) config.Job {
	j := config.Job{Name: "cooldown-job", Command: "echo hi"}
	if cooldown != "" {
		j.Options = map[string]string{"cooldown": cooldown}
	}
	return j
}

func TestCooldown_FirstRunAlwaysExecutes(t *testing.T) {
	called := 0
	inner := runnerFunc(func(_ context.Context, job config.Job) runner.Result {
		called++
		return runner.Result{Job: job}
	})
	cr := NewCooldownRunner(inner, 5*time.Minute)
	cr.Run(context.Background(), cooldownJob(""))
	if called != 1 {
		t.Fatalf("expected inner called once, got %d", called)
	}
}

func TestCooldown_SecondRunSkippedWithinCooldown(t *testing.T) {
	called := 0
	inner := runnerFunc(func(_ context.Context, job config.Job) runner.Result {
		called++
		return runner.Result{Job: job}
	})
	cr := NewCooldownRunner(inner, 5*time.Minute)
	job := cooldownJob("5m")

	cr.Run(context.Background(), job)
	res := cr.Run(context.Background(), job)

	if called != 1 {
		t.Fatalf("expected inner called once, got %d", called)
	}
	if !res.Skipped {
		t.Error("expected second run to be skipped")
	}
}

func TestCooldown_RunsAfterCooldownExpires(t *testing.T) {
	called := 0
	inner := runnerFunc(func(_ context.Context, job config.Job) runner.Result {
		called++
		return runner.Result{Job: job}
	})
	cr := NewCooldownRunner(inner, 10*time.Millisecond)
	job := cooldownJob("10ms")

	cr.Run(context.Background(), job)
	time.Sleep(20 * time.Millisecond)
	cr.Run(context.Background(), job)

	if called != 2 {
		t.Fatalf("expected inner called twice, got %d", called)
	}
}

func TestCooldown_ZeroCooldownIsNoop(t *testing.T) {
	called := 0
	inner := runnerFunc(func(_ context.Context, job config.Job) runner.Result {
		called++
		return runner.Result{Job: job}
	})
	cr := NewCooldownRunner(inner, 0)
	job := cooldownJob("")

	cr.Run(context.Background(), job)
	cr.Run(context.Background(), job)

	if called != 2 {
		t.Fatalf("expected inner called twice, got %d", called)
	}
}

func TestCooldownFromJob_ParsesDuration(t *testing.T) {
	tests := []struct {
		val  string
		want time.Duration
	}{
		{"30s", 30 * time.Second},
		{"2m", 2 * time.Minute},
		{"", 0},
		{"invalid", 0},
		{"-5s", 0},
	}
	for _, tc := range tests {
		job := cooldownJob(tc.val)
		got := CooldownFromJob(job)
		if got != tc.want {
			t.Errorf("CooldownFromJob(%q) = %v, want %v", tc.val, got, tc.want)
		}
	}
}

func TestWrapWithCooldown_PassthroughWhenNoCooldown(t *testing.T) {
	inner := runnerFunc(func(_ context.Context, job config.Job) runner.Result {
		return runner.Result{Job: job}
	})
	job := cooldownJob("")
	wrapped := WrapWithCooldown(inner, job)
	if _, ok := wrapped.(*CooldownRunner); ok {
		t.Error("expected passthrough, got CooldownRunner")
	}
}
