package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cronwatch/internal/runner"
)

// RetryPolicy defines how a job should be retried on failure.
type RetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
}

// RetryRunner wraps a runner and retries failed jobs according to a policy.
type RetryRunner struct {
	r      *runner.Runner
	policy RetryPolicy
}

// NewRetryRunner creates a RetryRunner with the given policy.
func NewRetryRunner(r *runner.Runner, policy RetryPolicy) *RetryRunner {
	if policy.MaxAttempts < 1 {
		policy.MaxAttempts = 1
	}
	return &RetryRunner{r: r, policy: policy}
}

// Run executes the job, retrying on failure up to MaxAttempts times.
// It returns the last result regardless of success or failure.
func (rr *RetryRunner) Run(ctx context.Context, jobName, command string, timeout time.Duration) runner.Result {
	var result runner.Result
	for attempt := 1; attempt <= rr.policy.MaxAttempts; attempt++ {
		result = rr.r.Run(ctx, jobName, command, timeout)
		if result.ExitCode == 0 {
			if attempt > 1 {
				log.Printf("[cronwatch] job %q succeeded on attempt %d/%d", jobName, attempt, rr.policy.MaxAttempts)
			}
			return result
		}
		if attempt < rr.policy.MaxAttempts {
			log.Printf("[cronwatch] job %q failed (attempt %d/%d, exit %d), retrying in %s",
				jobName, attempt, rr.policy.MaxAttempts, result.ExitCode, rr.policy.Delay)
			select {
			case <-time.After(rr.policy.Delay):
			case <-ctx.Done():
				result.Error = fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
				return result
			}
		}
	}
	return result
}
