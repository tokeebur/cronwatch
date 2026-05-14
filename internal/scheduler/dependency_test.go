package scheduler

import (
	"context"
	"errors"
	"testing"

	"github.com/cronwatch/cronwatch/internal/runner"
)

type stubDepRunner struct {
	called bool
}

func (s *stubDepRunner) Run(_ context.Context, job runner.Job) runner.Result {
	s.called = true
	return runner.Result{Job: job, ExitCode: 0}
}

func depJob(name string, deps []string) runner.Job {
	meta := map[string]interface{}{}
	if len(deps) > 0 {
		meta["depends_on"] = deps
	}
	return runner.Job{Name: name, Command: "echo ok", Meta: meta}
}

func TestDependency_RunsWhenDepsSucceeded(t *testing.T) {
	store := NewResultStore()
	store.Set("setup", runner.Result{ExitCode: 0})

	stub := &stubDepRunner{}
	job := depJob("main", []string{"setup"})
	dr := NewDependencyRunner(stub, []string{"setup"}, store)

	r := dr.Run(context.Background(), job)
	if r.Err != nil {
		t.Fatalf("expected no error, got %v", r.Err)
	}
	if !stub.called {
		t.Fatal("expected inner runner to be called")
	}
}

func TestDependency_SkipsWhenDepNotRun(t *testing.T) {
	store := NewResultStore()
	stub := &stubDepRunner{}
	job := depJob("main", []string{"setup"})
	dr := NewDependencyRunner(stub, []string{"setup"}, store)

	r := dr.Run(context.Background(), job)
	if r.Err == nil {
		t.Fatal("expected error when dependency has not run")
	}
	if stub.called {
		t.Fatal("inner runner should not be called")
	}
}

func TestDependency_SkipsWhenDepFailed(t *testing.T) {
	store := NewResultStore()
	store.Set("setup", runner.Result{ExitCode: 1, Err: errors.New("failed")})

	stub := &stubDepRunner{}
	job := depJob("main", []string{"setup"})
	dr := NewDependencyRunner(stub, []string{"setup"}, store)

	r := dr.Run(context.Background(), job)
	if r.Err == nil {
		t.Fatal("expected error when dependency failed")
	}
	if stub.called {
		t.Fatal("inner runner should not be called")
	}
}

func TestWrapWithDependency_NoDeps(t *testing.T) {
	store := NewResultStore()
	stub := &stubDepRunner{}
	job := depJob("main", nil)

	wrapped := WrapWithDependency(stub, job, store)
	if _, ok := wrapped.(*DependencyRunner); ok {
		t.Fatal("should not wrap when no deps")
	}
}

func TestResultStore_SetAndGet(t *testing.T) {
	store := NewResultStore()
	store.Set("job1", runner.Result{ExitCode: 0})

	r, ok := store.Get("job1")
	if !ok {
		t.Fatal("expected result to be found")
	}
	if r.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d", r.ExitCode)
	}

	_, ok = store.Get("missing")
	if ok {
		t.Fatal("expected missing key to not be found")
	}
}
