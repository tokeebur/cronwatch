package scheduler

import (
	"errors"
	"testing"
	"time"
)

// stubAlerter is a test double that records calls and optionally returns an error.
type stubAlerter struct {
	called  int
	lastResult Result
	errToReturn error
}

func (s *stubAlerter) Alert(r Result) error {
	s.called++
	s.lastResult = r
	return s.errToReturn
}

func makeResult(job string, ok bool) Result {
	return Result{
		JobName:  job,
		Success:  ok,
		ExitCode: 0,
		Output:   "",
		Duration: time.Second,
	}
}

func TestMultiAlerter_CallsAllAlerters(t *testing.T) {
	a1 := &stubAlerter{}
	a2 := &stubAlerter{}
	ma := NewMultiAlerter(a1, a2)

	r := makeResult("backup", false)
	if err := ma.Alert(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a1.called != 1 || a2.called != 1 {
		t.Errorf("expected each alerter called once, got a1=%d a2=%d", a1.called, a2.called)
	}
	if a1.lastResult.JobName != "backup" {
		t.Errorf("unexpected job name: %s", a1.lastResult.JobName)
	}
}

func TestMultiAlerter_ReturnsErrorWhenOneFails(t *testing.T) {
	a1 := &stubAlerter{errToReturn: errors.New("smtp down")}
	a2 := &stubAlerter{}
	ma := NewMultiAlerter(a1, a2)

	err := ma.Alert(makeResult("sync", false))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if a2.called != 1 {
		t.Error("second alerter should still be called when first fails")
	}
}

func TestMultiAlerter_AggregatesMultipleErrors(t *testing.T) {
	a1 := &stubAlerter{errToReturn: errors.New("err1")}
	a2 := &stubAlerter{errToReturn: errors.New("err2")}
	ma := NewMultiAlerter(a1, a2)

	err := ma.Alert(makeResult("job", false))
	if err == nil {
		t.Fatal("expected combined error")
	}
	msg := err.Error()
	if msg == "err1" || msg == "err2" {
		t.Errorf("expected combined message, got: %s", msg)
	}
}

func TestMultiAlerter_NoAlerters(t *testing.T) {
	ma := NewMultiAlerter()
	if err := ma.Alert(makeResult("job", false)); err != nil {
		t.Fatalf("unexpected error with no alerters: %v", err)
	}
}
