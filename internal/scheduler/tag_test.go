package scheduler

import (
	"context"
	"strings"
	"testing"

	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

func tagJob(tags ...string) config.Job {
	return config.Job{
		Name:    "tag-job",
		Command: "echo hello",
		Tags:    tags,
	}
}

type stubTagRunner struct{ result runner.Result }

func (s *stubTagRunner) Run(_ context.Context, _ config.Job) runner.Result { return s.result }

func TestTaggedRunner_AppendsTags(t *testing.T) {
	inner := &stubTagRunner{result: runner.Result{JobName: "tag-job", Output: "hello", Success: true}}
	tr := NewTaggedRunner(inner, []string{"prod", "critical"})

	res := tr.Run(context.Background(), tagJob("prod", "critical"))

	if !strings.Contains(res.Output, "[tags: prod, critical]") {
		t.Errorf("expected tags in output, got: %q", res.Output)
	}
	if !strings.HasPrefix(res.Output, "hello") {
		t.Errorf("expected original output preserved, got: %q", res.Output)
	}
}

func TestTaggedRunner_NoTagsNoChange(t *testing.T) {
	inner := &stubTagRunner{result: runner.Result{JobName: "tag-job", Output: "hello", Success: true}}
	tr := NewTaggedRunner(inner, nil)

	res := tr.Run(context.Background(), tagJob())

	if res.Output != "hello" {
		t.Errorf("expected unchanged output, got: %q", res.Output)
	}
}

func TestTaggedRunner_EmptyOutputShowsOnlyTags(t *testing.T) {
	inner := &stubTagRunner{result: runner.Result{JobName: "tag-job", Output: "", Success: true}}
	tr := NewTaggedRunner(inner, []string{"batch"})

	res := tr.Run(context.Background(), tagJob("batch"))

	if res.Output != "[tags: batch]" {
		t.Errorf("unexpected output: %q", res.Output)
	}
}

func TestNormaliseTags_LowercaseAndDedup(t *testing.T) {
	out := normaliseTags([]string{"Prod", "prod", " CRITICAL ", ""})
	if len(out) != 2 {
		t.Fatalf("expected 2 unique tags, got %d: %v", len(out), out)
	}
	if out[0] != "prod" || out[1] != "critical" {
		t.Errorf("unexpected normalised tags: %v", out)
	}
}

func TestWrapWithTags_WrapsWhenTagsPresent(t *testing.T) {
	inner := &stubTagRunner{result: runner.Result{Output: "ok", Success: true}}
	job := tagJob("env:prod")
	wrapped := WrapWithTags(inner, job)

	if _, ok := wrapped.(*TaggedRunner); !ok {
		t.Error("expected WrapWithTags to return a *TaggedRunner")
	}
}

func TestWrapWithTags_PassthroughWhenNoTags(t *testing.T) {
	inner := &stubTagRunner{result: runner.Result{Output: "ok", Success: true}}
	job := tagJob()
	wrapped := WrapWithTags(inner, job)

	if _, ok := wrapped.(*TaggedRunner); ok {
		t.Error("expected WrapWithTags to return inner unchanged when no tags")
	}
}
