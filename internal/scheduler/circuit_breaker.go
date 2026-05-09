package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/user/cronwatch/internal/runner"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker wraps a Runner and stops executing jobs when too many
// consecutive failures occur, allowing recovery after a cooldown period.
type CircuitBreaker struct {
	runner      runner.Runner
	maxFailures int
	cooldown    time.Duration

	mu          sync.Mutex
	failures    int
	state       CircuitState
	openedAt    time.Time
}

// NewCircuitBreaker creates a CircuitBreaker wrapping r.
func NewCircuitBreaker(r runner.Runner, maxFailures int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		runner:      r,
		maxFailures: maxFailures,
		cooldown:    cooldown,
		state:       StateClosed,
	}
}

// Run executes the job unless the circuit is open.
func (cb *CircuitBreaker) Run(ctx context.Context, job runner.Job) runner.Result {
	cb.mu.Lock()
	switch cb.state {
	case StateOpen:
		if time.Since(cb.openedAt) >= cb.cooldown {
			cb.state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return runner.Result{
				Job:   job,
				Error: fmt.Errorf("circuit open: too many failures for job %q", job.Name),
			}
		}
	case StateHalfOpen, StateClosed:
		// proceed
	}
	cb.mu.Unlock()

	res := cb.runner.Run(ctx, job)

	cb.mu.Lock()
	defer cb.mu.Unlock()
	if res.Error != nil || res.ExitCode != 0 {
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
			cb.openedAt = time.Now()
		}
	} else {
		cb.failures = 0
		cb.state = StateClosed
	}
	return res
}

// State returns the current circuit state (safe for concurrent use).
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
