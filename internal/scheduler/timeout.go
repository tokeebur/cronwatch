package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

// TimeoutRunner wraps a Runner and enforces a per-job execution deadline.
// If the job exceeds the configured timeout, the context is cancelled and
// the result is marked as a failure.
type TimeoutRunner struct {
	inner   runner.Runner
	timeout time.Duration
}

// NewTimeoutRunner returns a TimeoutRunner that cancels execution after d.
// If d is zero or negative, the inner runner is returned unwrapped.
func NewTimeoutRunner(inner runner.Runner, d time.Duration) runner.Runner {
	if d <= 0 {
		return inner
	}
	return &TimeoutRunner{inner: inner, timeout: d}
}

// Run executes the job with a deadline derived from the configured timeout.
func (t *TimeoutRunner) Run(ctx context.Context, job config.Job) runner.Result {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	result := t.inner.Run(ctx, job)

	// If the context deadline was exceeded, override the error for clarity.
	if ctx.Err() == context.DeadlineExceeded && result.Error == nil {
		result.Error = fmt.Errorf("job exceeded timeout of %s", t.timeout)
		result.ExitCode = -1
	}

	return result
}

// TimeoutFromJob reads the timeout value from a job's configuration and
// returns a wrapped runner. If no timeout is set the runner is unchanged.
func TimeoutFromJob(inner runner.Runner, job config.Job) runner.Runner {
	if job.Timeout == "" {
		return inner
	}
	d, err := time.ParseDuration(job.Timeout)
	if err != nil || d <= 0 {
		return inner
	}
	return NewTimeoutRunner(inner, d)
}
