package scheduler

import (
	"fmt"
	"time"

	"github.com/cronwatch/internal/config"
)

// RetryPolicyFromJob builds a RetryPolicy from a job's config.
// It returns a default single-attempt policy when no retry config is present.
func RetryPolicyFromJob(job config.Job) (RetryPolicy, error) {
	if job.Retry == nil {
		return RetryPolicy{MaxAttempts: 1, Delay: 0}, nil
	}

	max := job.Retry.MaxAttempts
	if max < 1 {
		return RetryPolicy{}, fmt.Errorf("job %q: retry.max_attempts must be >= 1, got %d", job.Name, max)
	}

	delay, err := time.ParseDuration(job.Retry.Delay)
	if err != nil {
		return RetryPolicy{}, fmt.Errorf("job %q: invalid retry.delay %q: %w", job.Name, job.Retry.Delay, err)
	}

	return RetryPolicy{
		MaxAttempts: max,
		Delay:       delay,
	}, nil
}
