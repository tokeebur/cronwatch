package scheduler

import (
	"time"

	"github.com/cronwatch/cronwatch/internal/config"
)

// DefaultThrottleCooldown is used when no per-job cooldown is configured.
const DefaultThrottleCooldown = 30 * time.Minute

// ThrottleFromJob returns the alert cooldown duration for the given job.
// It reads the optional alert_cooldown field (expressed as a Go duration string
// such as "15m" or "1h") and falls back to DefaultThrottleCooldown when the
// field is absent or unparseable.
func ThrottleFromJob(job config.Job) time.Duration {
	if job.AlertCooldown == "" {
		return DefaultThrottleCooldown
	}
	d, err := time.ParseDuration(job.AlertCooldown)
	if err != nil || d <= 0 {
		return DefaultThrottleCooldown
	}
	return d
}
