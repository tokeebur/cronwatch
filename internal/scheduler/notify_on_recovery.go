package scheduler

import (
	"context"
	"fmt"

	"github.com/celrenheit/cronwatch/internal/runner"
)

// recoveryAlerter wraps an Alerter and fires an additional "recovered" alert
// the first time a job succeeds after one or more failures.
type recoveryAlerter struct {
	inner   Alerter
	failed  map[string]bool
}

// NewRecoveryAlerter returns an Alerter that forwards all alerts to inner and
// additionally sends a synthetic recovery alert when a job transitions from
// failing to succeeding.
func NewRecoveryAlerter(inner Alerter) Alerter {
	return &recoveryAlerter{
		inner:  inner,
		failed: make(map[string]bool),
	}
}

func (r *recoveryAlerter) Alert(ctx context.Context, result runner.Result) error {
	name := result.Job.Name

	if !result.Success {
		r.failed[name] = true
		return r.inner.Alert(ctx, result)
	}

	// Job succeeded — check if it was previously failing.
	if r.failed[name] {
		delete(r.failed, name)
		recovery := result
		recovery.Output = fmt.Sprintf("[RECOVERED] %s completed successfully after previous failures.\n%s",
			name, result.Output)
		if err := r.inner.Alert(ctx, recovery); err != nil {
			return err
		}
	}

	return nil
}

// RecoveryFromJob returns true when the job config enables recovery alerts.
// By default recovery alerts are enabled.
func RecoveryFromJob(job interface{ GetRecoveryAlert() *bool }) bool {
	if v := job.GetRecoveryAlert(); v != nil {
		return *v
	}
	return true
}
