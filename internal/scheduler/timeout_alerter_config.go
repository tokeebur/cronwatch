package scheduler

import "github.com/cronwatch/cronwatch/internal/config"

// TimeoutAlertPolicy controls how the TimeoutAlerter behaves when a job
// exceeds its deadline.
type TimeoutAlertPolicy struct {
	// Enabled controls whether the wrapper is applied at all.
	// When false the inner alerter is used directly.
	Enabled bool
	// Forward determines whether timed-out results are forwarded to the inner
	// alerter (true) or silently dropped (false).
	Forward bool
}

// TimeoutAlertFromJob derives a TimeoutAlertPolicy from a job's configuration.
// Defaults: enabled=true, forward=true — operators receive timeout alerts but
// can opt-out per job.
func TimeoutAlertFromJob(job config.Job) TimeoutAlertPolicy {
	p := TimeoutAlertPolicy{
		Enabled: true,
		Forward: true,
	}
	if v, ok := job.Options["timeout_alert_enabled"]; ok {
		if b, ok := v.(bool); ok {
			p.Enabled = b
		}
	}
	if v, ok := job.Options["timeout_alert_forward"]; ok {
		if b, ok := v.(bool); ok {
			p.Forward = b
		}
	}
	return p
}

// WrapWithTimeoutAlert conditionally wraps alerter with a TimeoutAlerter
// according to the job's policy.
func WrapWithTimeoutAlert(alerter Alerter, job config.Job) Alerter {
	p := TimeoutAlertFromJob(job)
	if !p.Enabled {
		return alerter
	}
	return NewTimeoutAlerter(alerter, p.Forward)
}
