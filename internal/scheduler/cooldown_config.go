package scheduler

import (
	"fmt"
	"time"

	"github.com/nwlnexus/cronwatch/internal/config"
)

// ValidateCooldownInterval returns an error if the cooldown option on the job
// is present but cannot be parsed as a valid positive duration.
func ValidateCooldownInterval(job config.Job) error {
	raw, ok := job.Options["cooldown"]
	if !ok || raw == "" {
		return nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("job %q: invalid cooldown %q: %w", job.Name, raw, err)
	}
	if d < 0 {
		return fmt.Errorf("job %q: cooldown must be non-negative, got %q", job.Name, raw)
	}
	return nil
}
