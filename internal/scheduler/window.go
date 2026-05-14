package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

// WindowRunner wraps a Runner and only allows execution within a configured
// time-of-day window. Runs attempted outside the window are skipped.
type WindowRunner struct {
	inner  runner.Runner
	start  timeOfDay
	end    timeOfDay
	nowFn  func() time.Time
}

type timeOfDay struct {
	hour, minute int
}

func (t timeOfDay) asTime(base time.Time) time.Time {
	return time.Date(base.Year(), base.Month(), base.Day(), t.hour, t.minute, 0, 0, base.Location())
}

// NewWindowRunner creates a WindowRunner that delegates to inner only when the
// current time falls within [start, end) (both in "HH:MM" format).
func NewWindowRunner(inner runner.Runner, start, end string) (*WindowRunner, error) {
	s, err := parseTimeOfDay(start)
	if err != nil {
		return nil, fmt.Errorf("window start: %w", err)
	}
	e, err := parseTimeOfDay(end)
	if err != nil {
		return nil, fmt.Errorf("window end: %w", err)
	}
	return &WindowRunner{
		inner: inner,
		start: s,
		end:   e,
		nowFn: time.Now,
	}, nil
}

func (w *WindowRunner) Run(ctx context.Context, job config.Job) runner.Result {
	now := w.nowFn()
	winStart := w.start.asTime(now)
	winEnd := w.end.asTime(now)

	inWindow := now.Equal(winStart) || (now.After(winStart) && now.Before(winEnd))
	if !inWindow {
		return runner.Result{
			JobName: job.Name,
			Skipped: true,
			Output:  fmt.Sprintf("skipped: outside allowed window %s–%s", w.start, w.end),
		}
	}
	return w.inner.Run(ctx, job)
}

func (t timeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.hour, t.minute)
}

func parseTimeOfDay(s string) (timeOfDay, error) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return timeOfDay{}, fmt.Errorf("invalid time %q: expected HH:MM", s)
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return timeOfDay{}, fmt.Errorf("time %q out of range", s)
	}
	return timeOfDay{hour: h, minute: m}, nil
}

// WindowFromJob wraps r with a WindowRunner when the job config specifies a
// window. Returns r unchanged when no window is configured.
func WindowFromJob(r runner.Runner, job config.Job) (runner.Runner, error) {
	if job.Window.Start == "" && job.Window.End == "" {
		return r, nil
	}
	return NewWindowRunner(r, job.Window.Start, job.Window.End)
}
