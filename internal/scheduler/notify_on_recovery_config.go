package scheduler

import (
	"github.com/celrenheit/cronwatch/internal/config"
)

// WrapWithRecovery conditionally wraps alerter with a recoveryAlerter based
// on the job configuration. If the job does not define a recovery_alert field
// the feature is enabled by default.
func WrapWithRecovery(alerter Alerter, job config.Job) Alerter {
	if job.RecoveryAlert != nil && !*job.RecoveryAlert {
		return alerter
	}
	return NewRecoveryAlerter(alerter)
}
