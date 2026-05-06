// Package scheduler ties together the runner and notifier, executing
// configured cron jobs on their defined intervals and alerting on failures.
package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"cronwatch/internal/config"
	"cronwatch/internal/notifier"
	"cronwatch/internal/runner"
)

// Scheduler manages periodic execution of all configured jobs.
type Scheduler struct {
	cfg      *config.Config
	cron     *cron.Cron
	runner   *runner.Runner
	notifier *notifier.Notifier
	mu       sync.Mutex
	entries  map[string]cron.EntryID
}

// New creates a Scheduler wired to the provided config, runner, and notifier.
func New(cfg *config.Config, r *runner.Runner, n *notifier.Notifier) *Scheduler {
	return &Scheduler{
		cfg:      cfg,
		cron:     cron.New(),
		runner:   r,
		notifier: n,
		entries:  make(map[string]cron.EntryID),
	}
}

// Start registers all jobs and begins the cron loop. It blocks until ctx is
// cancelled, then performs a graceful shutdown.
func (s *Scheduler) Start(ctx context.Context) error {
	for _, job := range s.cfg.Jobs {
		job := job // capture
		id, err := s.cron.AddFunc(job.Schedule, func() {
			s.execute(job)
		})
		if err != nil {
			return err
		}
		s.mu.Lock()
		s.entries[job.Name] = id
		s.mu.Unlock()
		log.Printf("scheduler: registered job %q (%s)", job.Name, job.Schedule)
	}

	s.cron.Start()
	<-ctx.Done()
	log.Println("scheduler: shutting down")
	stopCtx := s.cron.Stop()
	select {
	case <-stopCtx.Done():
	case <-time.After(10 * time.Second):
		log.Println("scheduler: timed out waiting for running jobs")
	}
	return nil
}

func (s *Scheduler) execute(job config.Job) {
	log.Printf("scheduler: running job %q", job.Name)
	result := s.runner.Run(job)
	if !result.Success {
		log.Printf("scheduler: job %q failed, sending alert", job.Name)
		if err := s.notifier.Send(result); err != nil {
			log.Printf("scheduler: failed to send alert for job %q: %v", job.Name, err)
		}
	}
}
