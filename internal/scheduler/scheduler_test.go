package scheduler_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"cronwatch/internal/config"
	"cronwatch/internal/runner"
	"cronwatch/internal/scheduler"
)

// stubNotifier records how many alerts were sent.
type stubNotifier struct {
	count int32
}

func (s *stubNotifier) Send(_ runner.Result) error {
	atomic.AddInt32(&s.count, 1)
	return nil
}

func makeConfig(schedule, cmd string) *config.Config {
	return &config.Config{
		Jobs: []config.Job{
			{Name: "test-job", Schedule: schedule, Command: cmd, TimeoutSecs: 5},
		},
	}
}

func TestScheduler_RunsJob(t *testing.T) {
	cfg := makeConfig("@every 100ms", "echo hello")
	r := runner.New()
	n := &stubNotifier{}

	// scheduler.New accepts a notifier interface; adapt via thin wrapper.
	sched := scheduler.New(cfg, r, scheduler.NotifierFunc(n.Send))

	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- sched.Start(ctx) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	}
}

func TestScheduler_AlertOnFailure(t *testing.T) {
	cfg := makeConfig("@every 100ms", "false") // exits non-zero
	r := runner.New()
	n := &stubNotifier{}

	sched := scheduler.New(cfg, r, scheduler.NotifierFunc(n.Send))

	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- sched.Start(ctx) }()
	<-done

	if atomic.LoadInt32(&n.count) == 0 {
		t.Error("expected at least one alert to be sent for failing job")
	}
}

func TestScheduler_NoAlertOnSuccess(t *testing.T) {
	cfg := makeConfig("@every 100ms", "true")
	r := runner.New()
	n := &stubNotifier{}

	sched := scheduler.New(cfg, r, scheduler.NotifierFunc(n.Send))

	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- sched.Start(ctx) }()
	<-done

	if atomic.LoadInt32(&n.count) != 0 {
		t.Errorf("expected no alerts for successful job, got %d", n.count)
	}
}
