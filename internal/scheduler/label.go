package scheduler

import (
	"context"
	"fmt"

	"github.com/cronwatch/cronwatch/internal/runner"
)

// LabeledRunner wraps a Runner and injects structured metadata (labels) into
// the result's output prefix so downstream alerters and history entries carry
// consistent context without each component having to re-derive it.
type LabeledRunner struct {
	inner  Runner
	labels map[string]string
}

// NewLabeledRunner returns a LabeledRunner that delegates to inner and
// annotates every result with the supplied key/value labels.
func NewLabeledRunner(inner Runner, labels map[string]string) *LabeledRunner {
	copy := make(map[string]string, len(labels))
	for k, v := range labels {
		copy[k] = v
	}
	return &LabeledRunner{inner: inner, labels: copy}
}

// Run executes the inner runner and prepends a label header to the output.
func (l *LabeledRunner) Run(ctx context.Context, job runner.Job) (runner.Result, error) {
	res, err := l.inner.Run(ctx, job)
	if len(l.labels) == 0 {
		return res, err
	}
	header := l.buildHeader()
	res.Output = header + res.Output
	return res, err
}

func (l *LabeledRunner) buildHeader() string {
	s := "[labels]"
	for k, v := range l.labels {
		s += fmt.Sprintf(" %s=%s", k, v)
	}
	return s + "\n"
}

// LabelsFromJob extracts the optional labels map from a job config, returning
// nil when the job defines no labels.
func LabelsFromJob(job runner.Job) map[string]string {
	if len(job.Labels) == 0 {
		return nil
	}
	return job.Labels
}

// WrapWithLabels wraps inner with a LabeledRunner when the job defines labels,
// otherwise it returns inner unchanged.
func WrapWithLabels(inner Runner, job runner.Job) Runner {
	labels := LabelsFromJob(job)
	if len(labels) == 0 {
		return inner
	}
	return NewLabeledRunner(inner, labels)
}
