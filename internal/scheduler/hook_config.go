package scheduler

import (
	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

// HooksFromJob returns the pre/post hook commands declared in the job
// configuration, or empty strings when none are set.
func HooksFromJob(job config.Job) (pre, post string) {
	if job.Hooks == nil {
		return "", ""
	}
	return job.Hooks.Pre, job.Hooks.Post
}

// WrapWithHooks wraps r with a HookRunner when the job declares at least
// one hook command.  If neither pre nor post is set, r is returned
// unchanged.
func WrapWithHooks(r runner.Runner, job config.Job) runner.Runner {
	pre, post := HooksFromJob(job)
	if pre == "" && post == "" {
		return r
	}
	return NewHookRunner(r, pre, post, nil)
}
