package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/nwlnexus/cronwatch/internal/config"
	"github.com/nwlnexus/cronwatch/internal/runner"
)

// CooldownRunner prevents a job from being re-executed until a minimum
// cooldown period has elapsed since its last completion.
type CooldownRunner struct {
	inner    runner.Runner
	cooldown time.Duration

	mu       sync.Mutex
	lastDone time.Time
}

// NewCooldownRunner wraps inner so that it will be skipped if called before
// cooldown has elapsed since the previous run completed.
func NewCooldownRunner(inner runner.Runner, cooldown time.Duration) *CooldownRunner {
	return &CooldownRunner{
		inner:    inner,
		cooldown: cooldown,
	}
}

func (c *CooldownRunner) Run(ctx context.Context, job config.Job) runner.Result {
	if c.cooldown <= 0 {
		return c.inner.Run(ctx, job)
	}

	c.mu.Lock()
	sinceLastDone := time.Since(c.lastDone)
	if !c.lastDone.IsZero() && sinceLastDone < c.cooldown {
		c.mu.Unlock()
		return runner.Result{
			Job:     job,
			Skipped: true,
			Output:  "skipped: cooldown period has not elapsed",
		}
	}
	c.mu.Unlock()

	res := c.inner.Run(ctx, job)

	c.mu.Lock()
	c.lastDone = time.Now()
	c.mu.Unlock()

	return res
}

// CooldownFromJob returns the cooldown duration configured for a job.
// Returns 0 if not set or invalid.
func CooldownFromJob(job config.Job) time.Duration {
	raw, ok := job.Options["cooldown"]
	if !ok || raw == "" {
		return 0
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d < 0 {
		return 0
	}
	return d
}

// WrapWithCooldown wraps inner with a CooldownRunner when the job defines a
// cooldown option.
func WrapWithCooldown(inner runner.Runner, job config.Job) runner.Runner {
	d := CooldownFromJob(job)
	if d <= 0 {
		return inner
	}
	return NewCooldownRunner(inner, d)
}
