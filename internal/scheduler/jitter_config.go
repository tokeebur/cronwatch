package scheduler

import (
	"context"
	"time"

	"github.com/your-org/cronwatch/internal/config"
)

// Runner is the common interface for anything that can execute a cron job.
// It is declared here to avoid import cycles between scheduler sub-packages.
type Runner interface {
	Run(ctx context.Context, job config.Job) Result
}

// WrapWithJitter optionally wraps r with a JitterRunner based on the job's
// Jitter field. If the job has no jitter configured the original runner is
// returned unchanged, avoiding unnecessary allocation.
func WrapWithJitter(r Runner, job config.Job) Runner {
	d := JitterFromJob(job)
	if d <= 0 {
		return r
	}
	return NewJitterRunner(r, d)
}

// defaultJitter is used when a job omits the jitter field but the global
// config specifies a default. Zero means disabled.
const defaultJitter = 0 * time.Second
