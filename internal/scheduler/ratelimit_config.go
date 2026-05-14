package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/user/cronwatch/internal/runner"
)

// WrapWithRateLimit inspects the job metadata and, when a valid
// rate_limit_interval is configured, wraps the provided runner with a
// RateLimitedRunner. If no interval is set the original runner is returned
// unchanged. Configuration errors are returned so the caller can decide
// whether to abort startup.
func WrapWithRateLimit(r runner.Runner, job runner.Job) (runner.Runner, error) {
	interval, err := RateLimitFromJob(job)
	if err != nil {
		return nil, fmt.Errorf("ratelimit config: %w", err)
	}
	if interval == 0 {
		return r, nil
	}
	log.Printf("[cronwatch] job %q: rate-limit enabled (interval=%s)", job.Name, interval)
	return NewRateLimitedRunner(r, interval), nil
}

// DefaultRateLimitInterval is the minimum sensible interval. Values below
// this threshold are accepted but will log a warning.
const DefaultRateLimitInterval = 10 * time.Second

// ValidateRateLimitInterval warns when the configured interval is very small
// but does not treat it as an error, allowing operators to use short intervals
// in test environments.
func ValidateRateLimitInterval(job runner.Job) {
	interval, err := RateLimitFromJob(job)
	if err != nil || interval == 0 {
		return
	}
	if interval < DefaultRateLimitInterval {
		log.Printf("[cronwatch] warning: job %q has a very short rate_limit_interval (%s); consider >= %s",
			job.Name, interval, DefaultRateLimitInterval)
	}
}
