package runner

import (
	"context"
	"os/exec"
	"time"
)

// Result holds the outcome of a cron job execution.
type Result struct {
	JobName  string
	Command  string
	ExitCode int
	Output   string
	Duration time.Duration
	Err      error
}

// Runner executes shell commands and returns results.
type Runner struct {
	Timeout time.Duration
}

// New creates a Runner with the given timeout.
func New(timeout time.Duration) *Runner {
	return &Runner{Timeout: timeout}
}

// Run executes the given command string via sh -c and returns a Result.
func (r *Runner) Run(ctx context.Context, jobName, command string) Result {
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	out, err := cmd.CombinedOutput()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			err = nil // exit code captured; not a runner error
		}
	}

	return Result{
		JobName:  jobName,
		Command:  command,
		ExitCode: exitCode,
		Output:   string(out),
		Duration: duration,
		Err:      err,
	}
}
