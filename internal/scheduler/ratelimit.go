package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/user/cronwatch/internal/runner"
)

// RateLimitedRunner wraps a Runner and enforces a minimum interval between
// successive executions of the same job, dropping runs that arrive too soon.
type RateLimitedRunner struct {
	inner    runner.Runner
	interval time.Duration

	mu      sync.Mutex
	lastRun time.Time
}

// NewRateLimitedRunner returns a Runner that will skip execution if the
// previous run finished less than interval ago. Pass interval == 0 to
// disable rate-limiting (the inner runner is called unconditionally).
func NewRateLimitedRunner(inner runner.Runner, interval time.Duration) *RateLimitedRunner {
	return &RateLimitedRunner{
		inner:    inner,
		interval: interval,
	}
}

// Run executes the job unless it was run too recently, in which case it
// returns a skipped result with a descriptive message.
func (r *RateLimitedRunner) Run(ctx context.Context, job runner.Job) runner.Result {
	if r.interval > 0 {
		r.mu.Lock()
		sinceLastRun := time.Since(r.lastRun)
		if r.lastRun != (time.Time{}) && sinceLastRun < r.interval {
			r.mu.Unlock()
			return runner.Result{
				Job:    job,
				Output: fmt.Sprintf("[cronwatch] run skipped: rate-limited (next allowed in %s)", (r.interval - sinceLastRun).Round(time.Second)),
				Err:    nil,
				Skipped: true,
			}
		}
		r.lastRun = time.Now()
		r.mu.Unlock()
	}
	return r.inner.Run(ctx, job)
}

// RateLimitFromJob reads the rate_limit_interval field from a job config and
// returns the parsed duration. Returns 0 and no error when the field is absent.
func RateLimitFromJob(j runner.Job) (time.Duration, error) {
	raw, ok := j.Meta["rate_limit_interval"]
	if !ok || raw == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("job %q: invalid rate_limit_interval %q: %w", j.Name, raw, err)
	}
	if d < 0 {
		return 0, fmt.Errorf("job %q: rate_limit_interval must be non-negative", j.Name)
	}
	return d, nil
}
