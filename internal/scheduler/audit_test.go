package scheduler

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/cronwatch/cronwatch/internal/runner"
)

func newAuditLogger(buf *bytes.Buffer) *slog.Logger {
	h := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(h)
}

func auditJob(name string) runner.Job {
	return runner.Job{Name: name, Command: "echo hi"}
}

type staticAuditRunner struct{ res runner.Result }

func (s *staticAuditRunner) Run(_ context.Context, _ runner.Job) runner.Result { return s.res }

func TestAuditRunner_LogsOnSuccess(t *testing.T) {
	var buf bytes.Buffer
	inner := &staticAuditRunner{res: runner.Result{Success: true, ExitCode: 0, Output: "ok"}}
	ar := NewAuditRunner(inner, newAuditLogger(&buf))

	ar.Run(context.Background(), auditJob("myjob"))

	log := buf.String()
	if !strings.Contains(log, "myjob") {
		t.Errorf("expected job name in log, got: %s", log)
	}
	if !strings.Contains(log, "success=true") {
		t.Errorf("expected success=true in log, got: %s", log)
	}
}

func TestAuditRunner_LogsOnFailure(t *testing.T) {
	var buf bytes.Buffer
	inner := &staticAuditRunner{res: runner.Result{Success: false, ExitCode: 1, Output: "fail"}}
	ar := NewAuditRunner(inner, newAuditLogger(&buf))

	ar.Run(context.Background(), auditJob("badjob"))

	log := buf.String()
	if !strings.Contains(log, "success=false") {
		t.Errorf("expected success=false in log, got: %s", log)
	}
	if !strings.Contains(log, "exit_code=1") {
		t.Errorf("expected exit_code=1 in log, got: %s", log)
	}
}

func TestAuditRunner_TruncatesLongOutput(t *testing.T) {
	var buf bytes.Buffer
	long := strings.Repeat("x", 600)
	inner := &staticAuditRunner{res: runner.Result{Success: true, Output: long}}
	ar := NewAuditRunner(inner, newAuditLogger(&buf))

	ar.Run(context.Background(), auditJob("verbosejob"))

	if !strings.Contains(buf.String(), "truncated") {
		t.Error("expected truncation marker in audit log")
	}
}

func TestAuditFromJob_DefaultsTrue(t *testing.T) {
	j := runner.Job{Name: "j", Meta: map[string]string{}}
	if !AuditFromJob(j) {
		t.Error("expected audit enabled by default")
	}
}

func TestAuditFromJob_DisabledByMeta(t *testing.T) {
	j := runner.Job{Name: "j", Meta: map[string]string{"audit": "false"}}
	if AuditFromJob(j) {
		t.Error("expected audit disabled when meta audit=false")
	}
}

func TestWrapWithAudit_PassthroughWhenDisabled(t *testing.T) {
	var buf bytes.Buffer
	inner := &staticAuditRunner{res: runner.Result{Success: true}}
	j := runner.Job{Name: "j", Meta: map[string]string{"audit": "false"}}

	wrapped := WrapWithAudit(inner, j, newAuditLogger(&buf))
	wrapped.Run(context.Background(), j)

	if buf.Len() != 0 {
		t.Error("expected no log output when audit is disabled")
	}
}
