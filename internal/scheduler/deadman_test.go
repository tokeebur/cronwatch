package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/config"
)

type captureAlerter struct {
	mu      sync.Mutex
	results []RunResult
}

func (c *captureAlerter) Alert(_ context.Context, r RunResult) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = append(c.results, r)
	return nil
}

func (c *captureAlerter) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.results)
}

func TestDeadMan_SuccessResetsTimer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cap := &captureAlerter{}
	d := NewDeadManAlerter(ctx, cap, "myjob", 200*time.Millisecond)

	// Successful result should NOT forward to inner and should reset timer.
	err := d.Alert(ctx, RunResult{JobName: "myjob", ExitCode: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait longer than the original deadline; timer was reset so no fire yet.
	time.Sleep(150 * time.Millisecond)
	if cap.count() != 0 {
		t.Errorf("expected 0 alerts, got %d", cap.count())
	}
}

func TestDeadMan_FiresWhenNoSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cap := &captureAlerter{}
	_ = NewDeadManAlerter(ctx, cap, "myjob", 80*time.Millisecond)

	// Do not call Alert with a success — timer should fire.
	time.Sleep(200 * time.Millisecond)

	if cap.count() == 0 {
		t.Error("expected dead-man alert to fire, got none")
	}
	if cap.results[0].JobName != "myjob" {
		t.Errorf("expected job name myjob, got %s", cap.results[0].JobName)
	}
}

func TestDeadMan_FailureForwardedToInner(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cap := &captureAlerter{}
	d := NewDeadManAlerter(ctx, cap, "myjob", 10*time.Second)

	failResult := RunResult{JobName: "myjob", ExitCode: 1, Err: errorf("exit 1")}
	if err := d.Alert(ctx, failResult); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.count() != 1 {
		t.Errorf("expected 1 forwarded alert, got %d", cap.count())
	}
}

func TestDeadManFromJob_ParsesDuration(t *testing.T) {
	job := config.Job{DeadMan: "1h"}
	if got := DeadManFromJob(job); got != time.Hour {
		t.Errorf("expected 1h, got %v", got)
	}
}

func TestDeadManFromJob_EmptyReturnsZero(t *testing.T) {
	if got := DeadManFromJob(config.Job{}); got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

func errorf(s string) error {
	return fmt.Errorf("%s", s)
}
