package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/user/cronwatch/internal/config"
)

// DeadManAlerter wraps an Alerter and fires if a job has not succeeded
// within the configured deadline window. It is reset each time the
// underlying job completes successfully.
type DeadManAlerter struct {
	mu       sync.Mutex
	inner    Alerter
	deadline time.Duration
	lastOK   time.Time
	timer    *time.Timer
	jobName  string
}

// NewDeadManAlerter creates a DeadManAlerter that fires inner if no
// successful run is observed within deadline. A background goroutine
// is started; it is stopped when ctx is cancelled.
func NewDeadManAlerter(ctx context.Context, inner Alerter, jobName string, deadline time.Duration) *DeadManAlerter {
	d := &DeadManAlerter{
		inner:    inner,
		deadline: deadline,
		jobName:  jobName,
		lastOK:   time.Now(),
	}
	d.timer = time.AfterFunc(deadline, func() {
		d.mu.Lock()
		sinceOK := time.Since(d.lastOK)
		d.mu.Unlock()

		result := RunResult{
			JobName: jobName,
			ExitCode: -1,
			Output:   fmt.Sprintf("dead-man switch fired: no successful run in %s (last ok: %s ago)", deadline, sinceOK.Round(time.Second)),
			Err:      fmt.Errorf("dead-man switch: job %q has not succeeded within %s", jobName, deadline),
		}
		_ = inner.Alert(ctx, result)
	})

	go func() {
		<-ctx.Done()
		d.timer.Stop()
	}()

	return d
}

// Alert forwards the result to the inner alerter. On success it resets
// the dead-man timer.
func (d *DeadManAlerter) Alert(ctx context.Context, result RunResult) error {
	if result.Err == nil && result.ExitCode == 0 {
		d.mu.Lock()
		d.lastOK = time.Now()
		d.timer.Reset(d.deadline)
		d.mu.Unlock()
		return nil
	}
	return d.inner.Alert(ctx, result)
}

// DeadManFromJob returns the dead-man deadline for a job, or zero if
// the feature is not configured.
func DeadManFromJob(job config.Job) time.Duration {
	if job.DeadMan == "" {
		return 0
	}
	d, err := time.ParseDuration(job.DeadMan)
	if err != nil || d <= 0 {
		return 0
	}
	return d
}
