package scheduler

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/cronwatch/cronwatch/internal/runner"
)

// HookRunner executes optional pre/post shell commands around a job.
// If the pre-hook fails the inner runner is skipped and the error is
// surfaced as a job failure.  Post-hook failures are logged but do not
// override an otherwise successful result.
type HookRunner struct {
	inner   runner.Runner
	preCmd  string
	postCmd string
	execFn  func(ctx context.Context, cmd string) error
}

// NewHookRunner wraps inner with optional pre/post command hooks.
// Pass an empty string to skip a hook.  execFn may be nil; the default
// implementation runs each hook via "sh -c <cmd>".
func NewHookRunner(inner runner.Runner, preCmd, postCmd string, execFn func(context.Context, string) error) *HookRunner {
	if execFn == nil {
		execFn = shellExec
	}
	return &HookRunner{
		inner:   inner,
		preCmd:  preCmd,
		postCmd: postCmd,
		execFn:  execFn,
	}
}

// Run executes the pre-hook (if any), delegates to the inner runner, then
// executes the post-hook (if any) regardless of the inner result.
func (h *HookRunner) Run(ctx context.Context, job runner.Job) runner.Result {
	if h.preCmd != "" {
		if err := h.execFn(ctx, h.preCmd); err != nil {
			return runner.Result{
				Job:    job,
				Err:    fmt.Errorf("pre-hook failed: %w", err),
				Output: fmt.Sprintf("pre-hook %q: %v", h.preCmd, err),
			}
		}
	}

	res := h.inner.Run(ctx, job)

	if h.postCmd != "" {
		if err := h.execFn(ctx, h.postCmd); err != nil {
			log.Printf("[cronwatch] post-hook for job %q failed (ignored): %v", job.Name, err)
		}
	}

	return res
}

// shellExec is the default hook executor: runs cmd via "sh -c".
func shellExec(ctx context.Context, cmd string) error {
	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}
