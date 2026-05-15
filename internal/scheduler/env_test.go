package scheduler

import (
	"context"
	"testing"

	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

// captureEnvRunner records the environment injected via context.
type captureEnvRunner struct {
	captured []string
}

func (c *captureEnvRunner) Run(ctx context.Context, job config.Job) (runner.Result, error) {
	c.captured = EnvFromContext(ctx)
	return runner.Result{Job: job, ExitCode: 0}, nil
}

func TestEnvRunner_InjectsVariables(t *testing.T) {
	cap := &captureEnvRunner{}
	job := config.Job{Name: "j", Command: "echo"}
	extra := []string{"FOO=bar", "BAZ=qux"}

	er := NewEnvRunner(cap, extra)
	_, err := er.Run(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	env := cap.captured
	if env == nil {
		t.Fatal("expected env in context, got nil")
	}
	if !containsKV(env, "FOO", "bar") {
		t.Errorf("expected FOO=bar in env, got %v", env)
	}
	if !containsKV(env, "BAZ", "qux") {
		t.Errorf("expected BAZ=qux in env, got %v", env)
	}
}

func TestEnvRunner_OverridesExistingKey(t *testing.T) {
	cap := &captureEnvRunner{}
	er := NewEnvRunner(cap, []string{"PATH=/custom"})
	_, _ = er.Run(context.Background(), config.Job{Name: "j", Command: "echo"})

	env := cap.captured
	if !containsKV(env, "PATH", "/custom") {
		t.Errorf("expected PATH=/custom, got %v", env)
	}
	// Ensure PATH appears only once.
	count := 0
	for _, kv := range env {
		if envKey_(kv) == "PATH" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected PATH exactly once, found %d times", count)
	}
}

func TestWrapWithEnv_PassthroughWhenNoEnv(t *testing.T) {
	cap := &captureEnvRunner{}
	job := config.Job{Name: "j", Command: "echo"}
	wrapped := WrapWithEnv(cap, job)
	if wrapped != runner.Runner(cap) {
		t.Error("expected passthrough when job has no env vars")
	}
}

func TestWrapWithEnv_WrapsWhenEnvPresent(t *testing.T) {
	cap := &captureEnvRunner{}
	job := config.Job{Name: "j", Command: "echo", Env: map[string]string{"X": "1"}}
	wrapped := WrapWithEnv(cap, job)
	if _, ok := wrapped.(*EnvRunner); !ok {
		t.Error("expected EnvRunner wrapper when job has env vars")
	}
}

func TestEnvFromContext_NilWhenNotSet(t *testing.T) {
	if got := EnvFromContext(context.Background()); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// containsKV reports whether KEY=VALUE is present in the slice.
func containsKV(env []string, key, value string) bool {
	target := key + "=" + value
	for _, kv := range env {
		if kv == target {
			return true
		}
	}
	return false
}
