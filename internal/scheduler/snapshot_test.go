package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/runner"
)

// stubRunner returns a fixed Result.
type stubSnapshotRunner struct {
	result runner.Result
}

func (s *stubSnapshotRunner) Run(_ context.Context) runner.Result { return s.result }

func TestSnapshotRunner_NoResultBeforeRun(t *testing.T) {
	sr := NewSnapshotRunner("job1", &stubSnapshotRunner{})
	_, ok := sr.Latest()
	if ok {
		t.Fatal("expected no result before first run")
	}
}

func TestSnapshotRunner_CapturesResult(t *testing.T) {
	want := runner.Result{Job: "job1", ExitCode: 0, Output: "hello", FinishedAt: time.Now()}
	sr := NewSnapshotRunner("job1", &stubSnapshotRunner{result: want})
	sr.Run(context.Background())
	got, ok := sr.Latest()
	if !ok {
		t.Fatal("expected a result after run")
	}
	if got.Output != want.Output {
		t.Errorf("output: got %q, want %q", got.Output, want.Output)
	}
	if got.ExitCode != want.ExitCode {
		t.Errorf("exit code: got %d, want %d", got.ExitCode, want.ExitCode)
	}
}

func TestSnapshotRunner_UpdatesOnSubsequentRun(t *testing.T) {
	first := runner.Result{Job: "job1", ExitCode: 0, Output: "first"}
	second := runner.Result{Job: "job1", ExitCode: 1, Output: "second"}
	stub := &stubSnapshotRunner{result: first}
	sr := NewSnapshotRunner("job1", stub)
	sr.Run(context.Background())
	stub.result = second
	sr.Run(context.Background())
	got, _ := sr.Latest()
	if got.Output != second.Output {
		t.Errorf("expected updated output %q, got %q", second.Output, got.Output)
	}
}

func TestSnapshotStore_AllSkipsUnrun(t *testing.T) {
	store := NewSnapshotStore()
	store.Register(NewSnapshotRunner("idle", &stubSnapshotRunner{}))
	snaps := store.All()
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots for unrun jobs, got %d", len(snaps))
	}
}

func TestSnapshotStore_AllReturnsSingleEntry(t *testing.T) {
	store := NewSnapshotStore()
	sr := NewSnapshotRunner("myjob", &stubSnapshotRunner{
		result: runner.Result{Job: "myjob", ExitCode: 0, Output: "ok", FinishedAt: time.Now()},
	})
	store.Register(sr)
	sr.Run(context.Background())
	snaps := store.All()
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
	if snaps[0].Job != "myjob" {
		t.Errorf("expected job name 'myjob', got %q", snaps[0].Job)
	}
	if !snaps[0].Success {
		t.Error("expected success=true for exit code 0")
	}
}

func TestSnapshotStore_AllReportsFailure(t *testing.T) {
	store := NewSnapshotStore()
	sr := NewSnapshotRunner("failjob", &stubSnapshotRunner{
		result: runner.Result{Job: "failjob", ExitCode: 1, FinishedAt: time.Now()},
	})
	store.Register(sr)
	sr.Run(context.Background())
	snaps := store.All()
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
	if snaps[0].Success {
		t.Error("expected success=false for non-zero exit code")
	}
}
