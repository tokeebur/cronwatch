package scheduler

import (
	"strings"

	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/runner"
)

// ConcurrencyFromJob reads the concurrency_policy field from a job config and
// wraps the provided runner accordingly. If no policy is set the runner is
// returned unchanged.
func ConcurrencyFromJob(r runner.Runner, job config.Job) runner.Runner {
	policy := strings.ToLower(strings.TrimSpace(job.ConcurrencyPolicy))
	switch policy {
	case "skip":
		return NewConcurrencyRunner(r, PolicySkip)
	case "queue":
		return NewConcurrencyRunner(r, PolicyQueue)
	default:
		// "allow" or empty — no restriction
		return r
	}
}
