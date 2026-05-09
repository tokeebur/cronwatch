package scheduler

import (
	"sync"
	"time"
)

// ThrottledAlerter wraps an Alerter and suppresses repeated alerts for the
// same job within a configurable cooldown window. This prevents alert storms
// when a job fails repeatedly in quick succession.
type ThrottledAlerter struct {
	inner    Alerter
	cooldown time.Duration
	mu       sync.Mutex
	lastSent map[string]time.Time
}

// NewThrottledAlerter returns a ThrottledAlerter that delegates to inner but
// will not forward more than one alert per job within the cooldown duration.
func NewThrottledAlerter(inner Alerter, cooldown time.Duration) *ThrottledAlerter {
	return &ThrottledAlerter{
		inner:    inner,
		cooldown: cooldown,
		lastSent: make(map[string]time.Time),
	}
}

// Alert forwards the result to the inner Alerter only if no alert has been
// sent for this job within the cooldown window.
func (t *ThrottledAlerter) Alert(result Result) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if last, ok := t.lastSent[result.Job.Name]; ok {
		if now.Sub(last) < t.cooldown {
			return nil
		}
	}

	t.lastSent[result.Job.Name] = now
	return t.inner.Alert(result)
}

// Reset clears the throttle state for a specific job, allowing the next
// alert to be sent regardless of the cooldown window.
func (t *ThrottledAlerter) Reset(jobName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.lastSent, jobName)
}
