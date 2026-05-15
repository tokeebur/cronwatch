package scheduler

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cronwatch/cronwatch/internal/runner"
)

// AuditConfig holds global audit-logging settings parsed from the top-level
// config file.
type AuditConfig struct {
	// Enabled toggles audit logging globally. Defaults to true.
	Enabled bool `yaml:"enabled"`
	// Format is either "text" (default) or "json".
	Format string `yaml:"format"`
}

// DefaultAuditConfig returns an AuditConfig with sensible defaults.
func DefaultAuditConfig() AuditConfig {
	return AuditConfig{
		Enabled: true,
		Format:  "text",
	}
}

// Validate returns an error if the AuditConfig contains invalid values.
func (a AuditConfig) Validate() error {
	switch a.Format {
	case "text", "json", "":
		return nil
	default:
		return fmt.Errorf("audit: unknown format %q (want \"text\" or \"json\")", a.Format)
	}
}

// NewAuditLogger constructs a *slog.Logger from the AuditConfig, writing to
// stdout. Callers may replace this with any io.Writer.
func NewAuditLogger(cfg AuditConfig) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	var h slog.Handler
	if cfg.Format == "json" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(h)
}

// ApplyAudit wraps the provided runner with audit logging when both the global
// config and per-job meta agree that auditing is desired.
func ApplyAudit(inner runner.Runner, job runner.Job, cfg AuditConfig, logger *slog.Logger) runner.Runner {
	if !cfg.Enabled {
		return inner
	}
	return WrapWithAudit(inner, job, logger)
}
