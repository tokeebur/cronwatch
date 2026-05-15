package scheduler

import (
	"context"
	"errors"
	"testing"

	"github.com/cronwatch/cronwatch/internal/runner"
)

// stubRunner records calls and returns a preset result.
type stubHookRunner struct {
	called bool
	result runner.Result
}

func (s *stubHookRunner) Run(_ context.Context, job runner.Job) runner.Result {
	s.called = true
	s.result.Job = job
	return s.result
}

var hookJob = runner.Job{Name: "hook-job", Command: "true"}

func okExec(_ context.Context, _ string) error  { return nil }
func badExec(_ context.Context, _ string) error { return errors.New("hook error") }

func TestHookRunner_NoHooksPassthrough(t *testing.T) {
	inner := &stubHookRunner{}
	h := NewHookRunner(inner, "", "", okExec)
	res := h.Run(context.Background(), hookJob)
	if !inner.called {
		t.Fatal("expected inner runner to be called")
	}
	if res.Job.Name != hookJob.Name {
		t.Fatalf("job name mismatch: got %q", res.Job.Name)
	}
}

func TestHookRunner_PreHookFailureSkipsInner(t *testing.T) {
	inner := &stubHookRunner{}
	h := NewHookRunner(inner, "pre-cmd", "", badExec)
	res := h.Run(context.Background(), hookJob)
	if inner.called {
		t.Fatal("inner runner must not be called when pre-hook fails")
	}
	if res.Err == nil {
		t.Fatal("expected error from failed pre-hook")
	}
	if !errors.Is(res.Err, res.Err) {
		t.Fatal("error should be wrapped")
	}
}

func TestHookRunner_PostHookFailureDoesNotOverrideSuccess(t *testing.T) {
	inner := &stubHookRunner{}
	h := NewHookRunner(inner, "", "post-cmd", badExec)
	res := h.Run(context.Background(), hookJob)
	if !inner.called {
		t.Fatal("expected inner runner to be called")
	}
	// Post-hook failure must not surface as an error on the result.
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
}

func TestHookRunner_PreAndPostBothRun(t *testing.T) {
	calls := []string{}
	execFn := func(_ context.Context, cmd string) error {
		calls = append(calls, cmd)
		return nil
	}
	inner := &stubHookRunner{}
	h := NewHookRunner(inner, "pre", "post", execFn)
	h.Run(context.Background(), hookJob)

	if len(calls) != 2 {
		t.Fatalf("expected 2 hook calls, got %d", len(calls))
	}
	if calls[0] != "pre" || calls[1] != "post" {
		t.Fatalf("unexpected call order: %v", calls)
	}
}

func TestHookRunner_PostHookRunsEvenOnInnerFailure(t *testing.T) {
	postCalled := false
	execFn := func(_ context.Context, cmd string) error {
		if cmd == "post" {
			postCalled = true
		}
		return nil
	}
	inner := &stubHookRunner{result: runner.Result{Err: errors.New("inner failed")}}
	h := NewHookRunner(inner, "", "post", execFn)
	h.Run(context.Background(), hookJob)

	if !postCalled {
		t.Fatal("post-hook must run even when the inner job fails")
	}
}
