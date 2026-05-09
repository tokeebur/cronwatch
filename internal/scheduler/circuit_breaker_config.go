package scheduler

import (
	"time"

	"github.com/user/cronwatch/internal/config"
)

// CircuitBreakerPolicy holds parameters extracted from a job config.
type CircuitBreakerPolicy struct {
	Enabled     bool
	MaxFailures int
	Cooldown    time.Duration
}

const (
	defaultCBMaxFailures = 5
	defaultCBCooldown    = 5 * time.Minute
)

// CircuitBreakerFromJob extracts circuit-breaker settings from a job config.
// If the job does not define circuit-breaker settings the returned policy has
// Enabled == false.
func CircuitBreakerFromJob(j config.Job) CircuitBreakerPolicy {
	if j.CircuitBreaker == nil {
		return CircuitBreakerPolicy{}
	}
	p := CircuitBreakerPolicy{
		Enabled:     true,
		MaxFailures: defaultCBMaxFailures,
		Cooldown:    defaultCBCooldown,
	}
	if j.CircuitBreaker.MaxFailures > 0 {
		p.MaxFailures = j.CircuitBreaker.MaxFailures
	}
	if j.CircuitBreaker.CooldownSeconds > 0 {
		p.Cooldown = time.Duration(j.CircuitBreaker.CooldownSeconds) * time.Second
	}
	return p
}
