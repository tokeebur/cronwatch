package scheduler

import (
	"strconv"

	"github.com/yourorg/cronwatch/internal/config"
)

// RetryPolicy holds retry configuration derived from a job definition.
type RetryPolicy struct {
	MaxAttempts int
	Backoff     BackoffPolicy
}

// RetryPolicyFromJob builds a RetryPolicy from a job's Options map.
// Supported keys:
//
//	max_attempts      – total attempts including the first (default: 1, no retry)
//	backoff_strategy  – fixed | linear | exponential (default: exponential)
//	backoff_base      – duration string for base delay  (default: 1s)
//	backoff_max       – duration string for max delay   (default: 5m)
func RetryPolicyFromJob(job config.Job) RetryPolicy {
	policy := RetryPolicy{
		MaxAttempts: 1,
		Backoff:     DefaultBackoffPolicy(),
	}

	if job.Options == nil {
		return policy
	}

	if v, ok := job.Options["max_attempts"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			policy.MaxAttempts = n
		}
	}

	policy.Backoff = BackoffFromJob(job.Options)

	return policy
}
