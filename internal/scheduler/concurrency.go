package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/user/cronwatch/internal/runner"
)

// ConcurrencyPolicy controls what happens when a job is triggered while a
// previous instance is still running.
type ConcurrencyPolicy int

const (
	// PolicySkip drops the new execution attempt silently.
	PolicySkip ConcurrencyPolicy = iota
	// PolicyQueue waits for the running instance to finish before starting.
	PolicyQueue
)

// concurrentRunner wraps a Runner and enforces a concurrency policy.
type concurrentRunner struct {
	inner  runner.Runner
	policy ConcurrencyPolicy
	mu     sync.Mutex
	running bool
}

// NewConcurrencyRunner returns a Runner that enforces the given concurrency
// policy for the wrapped runner.
func NewConcurrencyRunner(inner runner.Runner, policy ConcurrencyPolicy) runner.Runner {
	return &concurrentRunner{inner: inner, policy: policy}
}

func (c *concurrentRunner) Run(ctx context.Context, job runner.Job) runner.Result {
	switch c.policy {
	case PolicySkip:
		c.mu.Lock()
		if c.running {
			c.mu.Unlock()
			return runner.Result{
				Job:    job,
				Err:    fmt.Errorf("skipped: previous instance of %q still running", job.Name),
				Skipped: true,
			}
		}
		c.running = true
		c.mu.Unlock()

		defer func() {
			c.mu.Lock()
			c.running = false
			c.mu.Unlock()
		}()
		return c.inner.Run(ctx, job)

	case PolicyQueue:
		c.mu.Lock()
		defer func() {
			c.running = false
			c.mu.Unlock()
		}()
		c.running = true
		return c.inner.Run(ctx, job)

	default:
		return c.inner.Run(ctx, job)
	}
}
