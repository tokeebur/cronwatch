package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/runner"
)

type stubRunner struct {
	calls   int
	results []runner.Result
}

func (s *stubRunner) Run(_ context.Context, job runner.Job) runner.Result {
	idx := s.calls
	if idx >= len(s.results) {
		idx = len(s.results) - 1
	}
	s.calls++
	res := s.results[idx]
	res.Job = job
	return res
}

func failResult(err error) runner.Result { return runner.Result{Error: err, ExitCode: 1} }
func okResult() runner.Result           { return runner.Result{ExitCode: 0} }

func TestCircuitBreaker_ClosedOnSuccess(t *testing.T) {
	sr := &stubRunner{results: []runner.Result{okResult()}}
	cb := NewCircuitBreaker(sr, 3, time.Minute)
	res := cb.Run(context.Background(), runner.Job{Name: "j"})
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if cb.State() != StateClosed {
		t.Fatalf("expected closed, got %v", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	err := errors.New("boom")
	sr := &stubRunner{results: []runner.Result{failResult(err)}}
	cb := NewCircuitBreaker(sr, 3, time.Minute)
	for i := 0; i < 3; i++ {
		cb.Run(context.Background(), runner.Job{Name: "j"})
	}
	if cb.State() != StateOpen {
		t.Fatalf("expected open after 3 failures, got %v", cb.State())
	}
}

func TestCircuitBreaker_BlocksWhenOpen(t *testing.T) {
	err := errors.New("boom")
	sr := &stubRunner{results: []runner.Result{failResult(err)}}
	cb := NewCircuitBreaker(sr, 2, time.Hour)
	for i := 0; i < 2; i++ {
		cb.Run(context.Background(), runner.Job{Name: "j"})
	}
	callsBefore := sr.calls
	res := cb.Run(context.Background(), runner.Job{Name: "j"})
	if res.Error == nil {
		t.Fatal("expected circuit-open error")
	}
	if sr.calls != callsBefore {
		t.Fatal("underlying runner should not be called when circuit is open")
	}
}

func TestCircuitBreaker_HalfOpenAfterCooldown(t *testing.T) {
	err := errors.New("boom")
	sr := &stubRunner{results: []runner.Result{failResult(err), okResult()}}
	cb := NewCircuitBreaker(sr, 1, 10*time.Millisecond)
	cb.Run(context.Background(), runner.Job{Name: "j"}) // trips breaker
	time.Sleep(20 * time.Millisecond)
	res := cb.Run(context.Background(), runner.Job{Name: "j"}) // half-open probe
	if res.Error != nil {
		t.Fatalf("expected success in half-open: %v", res.Error)
	}
	if cb.State() != StateClosed {
		t.Fatalf("expected closed after successful probe, got %v", cb.State())
	}
}
