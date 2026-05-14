package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/runner"
)

type blockingRunner struct {
	mu      sync.Mutex
	started chan struct{}
	unblock chan struct{}
}

func (b *blockingRunner) Run(ctx context.Context, job runner.Job) runner.Result {
	b.started <- struct{}{}
	<-b.unblock
	return runner.Result{Job: job}
}

func TestConcurrency_SkipDropsOverlappingRun(t *testing.T) {
	blk := &blockingRunner{
		started: make(chan struct{}, 1),
		unblock: make(chan struct{}),
	}
	cr := NewConcurrencyRunner(blk, PolicySkip)
	job := runner.Job{Name: "test", Command: "true"}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		cr.Run(context.Background(), job)
	}()

	<-blk.started

	result := cr.Run(context.Background(), job)
	if !result.Skipped {
		t.Errorf("expected Skipped=true, got false")
	}
	if result.Err == nil {
		t.Errorf("expected non-nil error for skipped run")
	}

	close(blk.unblock)
	wg.Wait()
}

func TestConcurrency_QueueSerializesRuns(t *testing.T) {
	order := make([]int, 0, 2)
	var mu sync.Mutex

	seq := &seqRunner{fn: func(i int) { mu.Lock(); order = append(order, i); mu.Unlock() }}
	cr := NewConcurrencyRunner(seq, PolicyQueue)
	job := runner.Job{Name: "q", Command: "true"}

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); cr.Run(context.Background(), job) }()
		time.Sleep(5 * time.Millisecond)
	}
	wg.Wait()

	if len(order) != 2 {
		t.Errorf("expected 2 runs, got %d", len(order))
	}
}

type seqRunner struct {
	mu sync.Mutex
	cnt int
	fn  func(int)
}

func (s *seqRunner) Run(_ context.Context, job runner.Job) runner.Result {
	s.mu.Lock()
	i := s.cnt
	s.cnt++
	s.mu.Unlock()
	s.fn(i)
	return runner.Result{Job: job}
}

func TestConcurrency_NoPolicyAllowsConcurrent(t *testing.T) {
	base := &seqRunner{fn: func(int) {}}
	r := ConcurrencyFromJob(base, jobWithPolicy(""))
	if r != base {
		t.Errorf("expected original runner when no policy set")
	}
}
