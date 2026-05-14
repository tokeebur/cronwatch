package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cronwatch/cronwatch/internal/runner"
)

// DependencyRunner wraps a Runner and blocks execution until all named
// dependency jobs have completed successfully within the current scheduling
// cycle. If a dependency has not run or failed, the job is skipped.
type DependencyRunner struct {
	inner  runner.Runner
	deps   []string
	lookup func(name string) (runner.Result, bool)
}

// ResultStore is a concurrency-safe store for job results within a cycle.
type ResultStore struct {
	mu      sync.RWMutex
	results map[string]runner.Result
}

func NewResultStore() *ResultStore {
	return &ResultStore{results: make(map[string]runner.Result)}
}

func (s *ResultStore) Set(name string, r runner.Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[name] = r
}

func (s *ResultStore) Get(name string) (runner.Result, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.results[name]
	return r, ok
}

// NewDependencyRunner creates a runner that checks deps before delegating.
func NewDependencyRunner(inner runner.Runner, deps []string, store *ResultStore) *DependencyRunner {
	return &DependencyRunner{
		inner:  inner,
		deps:   deps,
		lookup: store.Get,
	}
}

func (d *DependencyRunner) Run(ctx context.Context, job runner.Job) runner.Result {
	for _, dep := range d.deps {
		r, ok := d.lookup(dep)
		if !ok {
			return runner.Result{
				Job:      job,
				ExitCode: -1,
				Err:      fmt.Errorf("dependency %q has not run yet", dep),
				Duration: 0,
			}
		}
		if r.Err != nil || r.ExitCode != 0 {
			return runner.Result{
				Job:      job,
				ExitCode: -1,
				Err:      fmt.Errorf("dependency %q did not succeed (exit %d)", dep, r.ExitCode),
				Duration: 0,
			}
		}
	}
	return d.inner.Run(ctx, job)
}

// DepsFromJob returns the list of dependency job names from a job config.
func DepsFromJob(job runner.Job) []string {
	if v, ok := job.Meta["depends_on"]; ok {
		if deps, ok := v.([]string); ok {
			return deps
		}
	}
	return nil
}

// WrapWithDependency wraps inner with a DependencyRunner when deps are present.
func WrapWithDependency(inner runner.Runner, job runner.Job, store *ResultStore) runner.Runner {
	deps := DepsFromJob(job)
	if len(deps) == 0 {
		return inner
	}
	return NewDependencyRunner(inner, deps, store)
}

// unused import guard
var _ = time.Second
