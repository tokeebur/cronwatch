package scheduler

import (
	"testing"

	"github.com/user/cronwatch/internal/config"
	"github.com	/user/cronwatch/internal/runner"
)

func jobWithPolicy(p string) config.Job {
	return config.Job{Name: "j", Command: "true", ConcurrencyPolicy: p}
}

func TestConcurrencyFromJob_Skip(t *testing.T) {
	base := &seqRunner{fn: func(int) {}}
	r := ConcurrencyFromJob(base, jobWithPolicy("skip"))
	cr, ok := r.(*concurrentRunner)
	if !ok {
		t.Fatalf("expected *concurrentRunner, got %T", r)
	}
	if cr.policy != PolicySkip {
		t.Errorf("expected PolicySkip, got %v", cr.policy)
	}
}

func TestConcurrencyFromJob_Queue(t *testing.T) {
	base := &seqRunner{fn: func(int) {}}
	r := ConcurrencyFromJob(base, jobWithPolicy("queue"))
	cr, ok := r.(*concurrentRunner)
	if !ok {
		t.Fatalf("expected *concurrentRunner, got %T", r)
	}
	if cr.policy != PolicyQueue {
		t.Errorf("expected PolicyQueue, got %v", cr.policy)
	}
}

func TestConcurrencyFromJob_Allow(t *testing.T) {
	base := &seqRunner{fn: func(int) {}}
	r := ConcurrencyFromJob(base, jobWithPolicy("allow"))
	if _, ok := r.(*concurrentRunner); ok {
		t.Errorf("expected plain runner for policy=allow")
	}
}

func TestConcurrencyFromJob_CaseInsensitive(t *testing.T) {
	base := &seqRunner{fn: func(int) {}}
	r := ConcurrencyFromJob(base, jobWithPolicy("SKIP"))
	if _, ok := r.(*concurrentRunner); !ok {
		t.Errorf("expected *concurrentRunner for uppercase SKIP")
	}
}

func TestConcurrencyFromJob_UnknownPolicy(t *testing.T) {
	base := &seqRunner{fn: func(int) {}}
	r := ConcurrencyFromJob(base, jobWithPolicy("bogus"))
	if r != runner.Runner(base) {
		// unknown policy falls back to no wrapping
		t.Errorf("expected unwrapped runner for unknown policy")
	}
}
