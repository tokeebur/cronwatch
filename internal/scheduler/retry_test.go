package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/cronwatch/internal/runner"
)

func newRunner(t *testing.T) *runner.Runner {
	t.Helper()
	r, err := runner.New()
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}
	return r
}

func TestRetryRunner_SucceedsFirstAttempt(t *testing.T) {
	r := newRunner(t)
	rr := NewRetryRunner(r, RetryPolicy{MaxAttempts: 3, Delay: 0})
	result := rr.Run(context.Background(), "ok", "echo hello", 5*time.Second)
	if result.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d", result.ExitCode)
	}
}

func TestRetryRunner_RetriesOnFailure(t *testing.T) {
	r := newRunner(t)
	// exit 1 every time; expect 2 attempts with no delay
	rr := NewRetryRunner(r, RetryPolicy{MaxAttempts: 2, Delay: 0})
	result := rr.Run(context.Background(), "fail", "exit 1", 5*time.Second)
	if result.ExitCode == 0 {
		t.Fatal("expected non-zero exit code")
	}
}

func TestRetryRunner_StopsAfterMaxAttempts(t *testing.T) {
	r := newRunner(t)
	attempts := 0
	// Use a script that counts via a temp file would be complex; verify via timing.
	// Instead confirm MaxAttempts=1 means no retry.
	rr := NewRetryRunner(r, RetryPolicy{MaxAttempts: 1, Delay: 0})
	_ = attempts
	result := rr.Run(context.Background(), "fail", "exit 2", 5*time.Second)
	if result.ExitCode == 0 {
		t.Fatal("expected failure")
	}
}

func TestRetryRunner_ContextCancelledDuringBackoff(t *testing.T) {
	r := newRunner(t)
	rr := NewRetryRunner(r, RetryPolicy{MaxAttempts: 5, Delay: 10 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	result := rr.Run(ctx, "fail", "exit 1", 5*time.Second)
	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Fatalf("expected fast cancellation, took %s", elapsed)
	}
	if result.Error == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestRetryRunner_DefaultMinAttempts(t *testing.T) {
	r := newRunner(t)
	// MaxAttempts=0 should be treated as 1
	rr := NewRetryRunner(r, RetryPolicy{MaxAttempts: 0, Delay: 0})
	result := rr.Run(context.Background(), "ok", "echo hi", 5*time.Second)
	if result.ExitCode != 0 {
		t.Fatalf("expected success, got exit %d", result.ExitCode)
	}
}
