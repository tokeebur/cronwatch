package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cronwatch/cronwatch/internal/runner"
)

// AuditEvent records a structured log entry for every job execution attempt.
type AuditEvent struct {
	JobName   string
	StartedAt time.Time
	Duration  time.Duration
	Success   bool
	ExitCode  int
	Output    string
	Attempt   int
}

// AuditRunner wraps an inner Runner and emits a structured audit log after
// every execution regardless of outcome.
type AuditRunner struct {
	inner  runner.Runner
	logger *slog.Logger
}

// NewAuditRunner returns an AuditRunner that delegates to inner and writes
// one slog record per run using the provided logger.
func NewAuditRunner(inner runner.Runner, logger *slog.Logger) *AuditRunner {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuditRunner{inner: inner, logger: logger}
}

// Run executes the inner runner and emits an audit log entry.
func (a *AuditRunner) Run(ctx context.Context, job runner.Job) runner.Result {
	start := time.Now()
	res := a.inner.Run(ctx, job)
	dur := time.Since(start)

	level := slog.LevelInfo
	if !res.Success {
		level = slog.LevelWarn
	}

	a.logger.LogAttrs(ctx, level, "audit",
		slog.String("job", job.Name),
		slog.Time("started_at", start),
		slog.Duration("duration", dur),
		slog.Bool("success", res.Success),
		slog.Int("exit_code", res.ExitCode),
		slog.String("output", truncateAuditOutput(res.Output, 512)),
	)

	return res
}

func truncateAuditOutput(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return fmt.Sprintf("%s… [truncated %d bytes]", s[:max], len(s)-max)
}

// AuditFromJob returns true when audit logging is enabled for the job.
// By default audit logging is always on; jobs may opt out via audit: false.
func AuditFromJob(job runner.Job) bool {
	val, ok := job.Meta["audit"]
	if !ok {
		return true
	}
	return val != "false" && val != "0" && val != "no"
}

// WrapWithAudit wraps inner with an AuditRunner when audit is enabled for the
// job, otherwise returns inner unchanged.
func WrapWithAudit(inner runner.Runner, job runner.Job, logger *slog.Logger) runner.Runner {
	if !AuditFromJob(job) {
		return inner
	}
	return NewAuditRunner(inner, logger)
}
