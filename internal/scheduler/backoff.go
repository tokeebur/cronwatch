package scheduler

import (
	"math"
	"time"
)

// BackoffStrategy defines how retry delays are calculated.
type BackoffStrategy int

const (
	// BackoffFixed uses the same delay for every retry attempt.
	BackoffFixed BackoffStrategy = iota
	// BackoffLinear increases the delay linearly with each attempt.
	BackoffLinear
	// BackoffExponential doubles the delay with each attempt.
	BackoffExponential
)

// BackoffPolicy holds the configuration for a backoff strategy.
type BackoffPolicy struct {
	Strategy BackoffStrategy
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

// DefaultBackoffPolicy returns a sensible exponential backoff default.
func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{
		Strategy:  BackoffExponential,
		BaseDelay: 1 * time.Second,
		MaxDelay:  5 * time.Minute,
	}
}

// Delay returns the wait duration for the given attempt (0-indexed).
func (p BackoffPolicy) Delay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	var d time.Duration

	switch p.Strategy {
	case BackoffLinear:
		d = p.BaseDelay * time.Duration(attempt+1)
	case BackoffExponential:
		mult := math.Pow(2, float64(attempt))
		d = time.Duration(float64(p.BaseDelay) * mult)
	default: // BackoffFixed
		d = p.BaseDelay
	}

	if p.MaxDelay > 0 && d > p.MaxDelay {
		return p.MaxDelay
	}
	return d
}

// BackoffFromJob parses backoff configuration from a job's Options map.
// Supported keys: backoff_strategy (fixed|linear|exponential),
// backoff_base (duration string), backoff_max (duration string).
func BackoffFromJob(opts map[string]string) BackoffPolicy {
	policy := DefaultBackoffPolicy()

	if v, ok := opts["backoff_strategy"]; ok {
		switch v {
		case "fixed":
			policy.Strategy = BackoffFixed
		case "linear":
			policy.Strategy = BackoffLinear
		case "exponential":
			policy.Strategy = BackoffExponential
		}
	}

	if v, ok := opts["backoff_base"]; ok {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			policy.BaseDelay = d
		}
	}

	if v, ok := opts["backoff_max"]; ok {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			policy.MaxDelay = d
		}
	}

	return policy
}
