package scheduler

import (
	"context"
	"fmt"

	"github.com/cronwatch/cronwatch/internal/runner"
)

// TimeoutAlerter wraps an Alerter and suppresses alerts for jobs that were
// cancelled due to a timeout, optionally forwarding them with a distinct
// message so operators can distinguish timeout failures from real failures.
type TimeoutAlerter struct {
	inner   Alerter
	forward bool // if true, forward timeout alerts with an annotated message
}

// NewTimeoutAlerter creates a TimeoutAlerter.
// When forward is true the inner alerter is called with the result annotated
// to make the timeout origin clear. When false, timeout results are silently
// dropped.
func NewTimeoutAlerter(inner Alerter, forward bool) *TimeoutAlerter {
	return &TimeoutAlerter{inner: inner, forward: forward}
}

// Alert implements Alerter. It inspects the result to decide whether to
// forward or suppress the alert.
func (t *TimeoutAlerter) Alert(ctx context.Context, result runner.Result) error {
	if !result.TimedOut {
		return t.inner.Alert(ctx, result)
	}
	if !t.forward {
		return nil
	}
	// Annotate the output so the alert body is informative.
	annotated := result
	annotated.Output = fmt.Sprintf("[cronwatch] job timed out after %s\n%s",
		annotated.Duration, result.Output)
	return t.inner.Alert(ctx, annotated)
}
