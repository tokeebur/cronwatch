package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/runner"
)

type captureAlerter struct {
	called bool
	last   runner.Result
	err    error
}

func (c *captureAlerter) Alert(_ context.Context, r runner.Result) error {
	c.called = true
	c.last = r
	return c.err
}

func timedOutResult() runner.Result {
	return runner.Result{
		Job:      "myjob",
		TimedOut: true,
		Duration: 5 * time.Second,
		Output:   "partial output",
	}
}

func TestTimeoutAlerter_ForwardsNormalFailure(t *testing.T) {
	cap := &captureAlerter{}
	a := NewTimeoutAlerter(cap, false)
	r := runner.Result{Job: "myjob", ExitCode: 1}
	if err := a.Alert(context.Background(), r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cap.called {
		t.Fatal("expected inner alerter to be called for non-timeout failure")
	}
}

func TestTimeoutAlerter_SuppressesWhenForwardFalse(t *testing.T) {
	cap := &captureAlerter{}
	a := NewTimeoutAlerter(cap, false)
	if err := a.Alert(context.Background(), timedOutResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.called {
		t.Fatal("expected inner alerter to be suppressed for timeout result")
	}
}

func TestTimeoutAlerter_ForwardsAnnotatedWhenForwardTrue(t *testing.T) {
	cap := &captureAlerter{}
	a := NewTimeoutAlerter(cap, true)
	if err := a.Alert(context.Background(), timedOutResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cap.called {
		t.Fatal("expected inner alerter to be called")
	}
	if cap.last.Output == "partial output" {
		t.Error("expected output to be annotated with timeout message")
	}
}

func TestTimeoutAlerter_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("smtp down")
	cap := &captureAlerter{err: sentinel}
	a := NewTimeoutAlerter(cap, true)
	err := a.Alert(context.Background(), timedOutResult())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestWrapWithTimeoutAlert_DisabledPassthrough(t *testing.T) {
	cap := &captureAlerter{}
	job := makeJob("noop")
	job.Options = map[string]interface{}{"timeout_alert_enabled": false}
	wrapped := WrapWithTimeoutAlert(cap, job)
	// wrapped should be the original cap, not a TimeoutAlerter
	if _, ok := wrapped.(*TimeoutAlerter); ok {
		t.Fatal("expected passthrough when timeout_alert_enabled=false")
	}
}
