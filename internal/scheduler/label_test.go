package scheduler

import (
	"context"
	"strings"
	"testing"

	"github.com/cronwatch/cronwatch/internal/runner"
)

func labelJob(labels map[string]string) runner.Job {
	return runner.Job{
		Name:    "labeled-job",
		Command: "echo hello",
		Labels:  labels,
	}
}

func TestLabeledRunner_PrependsLabels(t *testing.T) {
	inner := &fakeRunner{result: runner.Result{Output: "hello\n"}}
	lr := NewLabeledRunner(inner, map[string]string{"env": "prod"})

	res, err := lr.Run(context.Background(), labelJob(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(res.Output, "[labels]") {
		t.Errorf("expected output to start with [labels], got: %q", res.Output)
	}
	if !strings.Contains(res.Output, "env=prod") {
		t.Errorf("expected label env=prod in output, got: %q", res.Output)
	}
	if !strings.Contains(res.Output, "hello") {
		t.Errorf("expected original output preserved, got: %q", res.Output)
	}
}

func TestLabeledRunner_NoLabelsNoPrefix(t *testing.T) {
	inner := &fakeRunner{result: runner.Result{Output: "clean\n"}}
	lr := NewLabeledRunner(inner, nil)

	res, err := lr.Run(context.Background(), labelJob(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Output != "clean\n" {
		t.Errorf("expected unmodified output, got: %q", res.Output)
	}
}

func TestWrapWithLabels_WrapsWhenLabelsPresent(t *testing.T) {
	inner := &fakeRunner{result: runner.Result{Output: "out"}}
	job := labelJob(map[string]string{"team": "platform"})
	wrapped := WrapWithLabels(inner, job)

	if _, ok := wrapped.(*LabeledRunner); !ok {
		t.Errorf("expected *LabeledRunner, got %T", wrapped)
	}
}

func TestWrapWithLabels_PassthroughWhenNoLabels(t *testing.T) {
	inner := &fakeRunner{result: runner.Result{Output: "out"}}
	job := labelJob(nil)
	wrapped := WrapWithLabels(inner, job)

	if _, ok := wrapped.(*fakeRunner); !ok {
		t.Errorf("expected passthrough *fakeRunner, got %T", wrapped)
	}
}

func TestLabelsFromJob_ReturnsNilWhenEmpty(t *testing.T) {
	job := labelJob(map[string]string{})
	if got := LabelsFromJob(job); got != nil {
		t.Errorf("expected nil for empty labels map, got %v", got)
	}
}

func TestLabeledRunner_IsolatesLabelCopy(t *testing.T) {
	orig := map[string]string{"k": "v"}
	inner := &fakeRunner{result: runner.Result{Output: "x"}}
	lr := NewLabeledRunner(inner, orig)
	orig["k"] = "mutated"

	res, _ := lr.Run(context.Background(), labelJob(nil))
	if strings.Contains(res.Output, "mutated") {
		t.Errorf("label mutation should not affect runner, got: %q", res.Output)
	}
}
