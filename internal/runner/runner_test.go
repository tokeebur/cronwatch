package runner_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/runner"
)

func TestRun_Success(t *testing.T) {
	r := runner.New(5 * time.Second)
	res := r.Run(context.Background(), "echo-job", "echo hello")

	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", res.ExitCode)
	}
	if !strings.Contains(res.Output, "hello") {
		t.Fatalf("expected output to contain 'hello', got %q", res.Output)
	}
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Duration <= 0 {
		t.Fatal("expected positive duration")
	}
}

func TestRun_NonZeroExit(t *testing.T) {
	r := runner.New(5 * time.Second)
	res := r.Run(context.Background(), "fail-job", "exit 42")

	if res.ExitCode != 42 {
		t.Fatalf("expected exit code 42, got %d", res.ExitCode)
	}
	if res.Err != nil {
		t.Fatalf("unexpected runner error: %v", res.Err)
	}
}

func TestRun_Timeout(t *testing.T) {
	r := runner.New(100 * time.Millisecond)
	res := r.Run(context.Background(), "slow-job", "sleep 10")

	if res.ExitCode == 0 {
		t.Fatal("expected non-zero exit code on timeout")
	}
}

func TestRun_CapturesOutput(t *testing.T) {
	r := runner.New(5 * time.Second)
	res := r.Run(context.Background(), "output-job", "echo stdout; echo stderr >&2")

	if !strings.Contains(res.Output, "stdout") {
		t.Errorf("expected stdout in output, got %q", res.Output)
	}
	if !strings.Contains(res.Output, "stderr") {
		t.Errorf("expected stderr in output, got %q", res.Output)
	}
}

func TestRun_JobNamePreserved(t *testing.T) {
	r := runner.New(5 * time.Second)
	res := r.Run(context.Background(), "my-job", "true")

	if res.JobName != "my-job" {
		t.Fatalf("expected job name 'my-job', got %q", res.JobName)
	}
}
