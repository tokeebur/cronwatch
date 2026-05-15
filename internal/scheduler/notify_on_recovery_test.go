package scheduler

import (
	"context"
	"strings"
	"testing"

	"github.com/celrenheit/cronwatch/internal/config"
	"github.com/celrenheit/cronwatch/internal/runner"
)

type captureAlerter struct {
	calls []runner.Result
}

func (c *captureAlerter) Alert(_ context.Context, r runner.Result) error {
	c.calls = append(c.calls, r)
	return nil
}

func recResult(name string, success bool) runner.Result {
	return runner.Result{
		Job:     config.Job{Name: name},
		Success: success,
		Output:  "some output",
	}
}

func TestRecoveryAlerter_NoAlertOnFirstSuccess(t *testing.T) {
	cap := &captureAlerter{}
	a := NewRecoveryAlerter(cap)

	_ = a.Alert(context.Background(), recResult("job1", true))

	if len(cap.calls) != 0 {
		t.Fatalf("expected no alerts, got %d", len(cap.calls))
	}
}

func TestRecoveryAlerter_AlertOnFailure(t *testing.T) {
	cap := &captureAlerter{}
	a := NewRecoveryAlerter(cap)

	_ = a.Alert(context.Background(), recResult("job1", false))

	if len(cap.calls) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(cap.calls))
	}
}

func TestRecoveryAlerter_SendsRecoveryAfterFailure(t *testing.T) {
	cap := &captureAlerter{}
	a := NewRecoveryAlerter(cap)

	_ = a.Alert(context.Background(), recResult("job1", false))
	_ = a.Alert(context.Background(), recResult("job1", true))

	if len(cap.calls) != 2 {
		t.Fatalf("expected 2 alerts (failure + recovery), got %d", len(cap.calls))
	}
	if !strings.Contains(cap.calls[1].Output, "[RECOVERED]") {
		t.Errorf("expected recovery message in output, got: %s", cap.calls[1].Output)
	}
}

func TestRecoveryAlerter_NoSecondRecoveryAlert(t *testing.T) {
	cap := &captureAlerter{}
	a := NewRecoveryAlerter(cap)

	_ = a.Alert(context.Background(), recResult("job1", false))
	_ = a.Alert(context.Background(), recResult("job1", true))
	_ = a.Alert(context.Background(), recResult("job1", true)) // second success

	if len(cap.calls) != 2 {
		t.Fatalf("expected exactly 2 alerts, got %d", len(cap.calls))
	}
}

func TestWrapWithRecovery_DisabledViaConfig(t *testing.T) {
	cap := &captureAlerter{}
	falseVal := false
	job := config.Job{Name: "j", RecoveryAlert: &falseVal}

	wrapped := WrapWithRecovery(cap, job)

	if _, ok := wrapped.(*recoveryAlerter); ok {
		t.Fatal("expected passthrough alerter when recovery_alert=false")
	}
}

func TestWrapWithRecovery_EnabledByDefault(t *testing.T) {
	cap := &captureAlerter{}
	job := config.Job{Name: "j"}

	wrapped := WrapWithRecovery(cap, job)

	if _, ok := wrapped.(*recoveryAlerter); !ok {
		t.Fatal("expected recoveryAlerter when no config provided")
	}
}
