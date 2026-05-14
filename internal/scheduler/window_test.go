package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

type stubWindowRunner struct {
	called bool
	result runner.Result
}

func (s *stubWindowRunner) Run(_ context.Context, job config.Job) runner.Result {
	s.called = true
	s.result.JobName = job.Name
	return s.result
}

func fixedNow(h, m int) func() time.Time {
	return func() time.Time {
		return time.Date(2024, 1, 15, h, m, 0, 0, time.UTC)
	}
}

func TestWindowRunner_ExecutesInsideWindow(t *testing.T) {
	stub := &stubWindowRunner{}
	w, err := NewWindowRunner(stub, "09:00", "17:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.nowFn = fixedNow(12, 30)

	res := w.Run(context.Background(), config.Job{Name: "job1"})
	if !stub.called {
		t.Error("expected inner runner to be called inside window")
	}
	if res.Skipped {
		t.Error("result should not be skipped inside window")
	}
}

func TestWindowRunner_SkipsOutsideWindow(t *testing.T) {
	stub := &stubWindowRunner{}
	w, err := NewWindowRunner(stub, "09:00", "17:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.nowFn = fixedNow(18, 0)

	res := w.Run(context.Background(), config.Job{Name: "job2"})
	if stub.called {
		t.Error("inner runner should not be called outside window")
	}
	if !res.Skipped {
		t.Error("result should be marked skipped outside window")
	}
	if res.JobName != "job2" {
		t.Errorf("expected JobName job2, got %q", res.JobName)
	}
}

func TestWindowRunner_ExecutesAtWindowBoundary(t *testing.T) {
	stub := &stubWindowRunner{}
	w, err := NewWindowRunner(stub, "09:00", "17:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.nowFn = fixedNow(9, 0) // exactly at start

	res := w.Run(context.Background(), config.Job{Name: "job3"})
	if !stub.called {
		t.Error("expected inner runner to be called at window start boundary")
	}
	if res.Skipped {
		t.Error("result should not be skipped at window start")
	}
}

func TestParseTimeOfDay_InvalidFormat(t *testing.T) {
	_, err := parseTimeOfDay("9am")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestParseTimeOfDay_OutOfRange(t *testing.T) {
	_, err := parseTimeOfDay("25:00")
	if err == nil {
		t.Error("expected error for out-of-range hour")
	}
}

func TestWindowFromJob_NoWindow_ReturnsUnchanged(t *testing.T) {
	stub := &stubWindowRunner{}
	job := config.Job{Name: "noop"}
	r, err := WindowFromJob(stub, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r != runner.Runner(stub) {
		t.Error("expected original runner returned when no window configured")
	}
}
