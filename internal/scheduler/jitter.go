package scheduler

import (
	"context"
	"math/rand"
	"time"

	"github.com/your-org/cronwatch/internal/config"
)

// JitterRunner wraps a Runner and adds a random delay before each execution.
// This helps spread out concurrent job starts when many jobs share the same
// cron schedule (e.g. every hour on the hour).
type JitterRunner struct {
	inner     Runner
	maxJitter time.Duration
	rng       *rand.Rand
}

// NewJitterRunner creates a JitterRunner that delays up to maxJitter before
// delegating to inner. If maxJitter is zero or negative the runner is a no-op
// wrapper and adds no delay.
func NewJitterRunner(inner Runner, maxJitter time.Duration) *JitterRunner {
	return &JitterRunner{
		inner:     inner,
		maxJitter: maxJitter,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Run sleeps for a random duration in [0, maxJitter) then delegates to the
// inner runner. The context deadline is respected; if the context expires
// during the jitter sleep the result reflects that cancellation.
func (j *JitterRunner) Run(ctx context.Context, job config.Job) Result {
	if j.maxJitter > 0 {
		delay := time.Duration(j.rng.Int63n(int64(j.maxJitter)))
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return Result{
				Job:   job,
				Error: ctx.Err(),
			}
		}
	}
	return j.inner.Run(ctx, job)
}

// JitterFromJob reads the jitter configuration from a job definition and
// returns the parsed duration. Returns 0 if not set or invalid.
func JitterFromJob(job config.Job) time.Duration {
	if job.Jitter == "" {
		return 0
	}
	d, err := time.ParseDuration(job.Jitter)
	if err != nil || d < 0 {
		return 0
	}
	return d
}

// NewJitterRunnerForJob is a convenience constructor that reads the jitter
// setting directly from a job definition and wraps the provided inner Runner.
// It is equivalent to calling NewJitterRunner(inner, JitterFromJob(job)).
func NewJitterRunnerForJob(inner Runner, job config.Job) *JitterRunner {
	return NewJitterRunner(inner, JitterFromJob(job))
}

// MaxJitter returns the maximum jitter duration configured for this runner.
// This is useful for logging or introspection without exposing the field
// directly.
func (j *JitterRunner) MaxJitter() time.Duration {
	return j.maxJitter
}
