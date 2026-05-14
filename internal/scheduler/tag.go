package scheduler

import (
	"context"
	"strings"

	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

// TaggedRunner wraps a Runner and injects job tags into the result metadata
// so that downstream components (alerters, history) can filter or annotate
// by tag.
type TaggedRunner struct {
	inner runner.Runner
	tags  []string
}

// NewTaggedRunner returns a TaggedRunner that decorates inner with the
// provided tags. Tags are normalised to lowercase and deduplicated.
func NewTaggedRunner(inner runner.Runner, tags []string) *TaggedRunner {
	norm := normaliseTags(tags)
	return &TaggedRunner{inner: inner, tags: norm}
}

// Run executes the inner runner and appends a "tags" annotation to the
// result output so that the tag list is visible in history and alerts.
func (t *TaggedRunner) Run(ctx context.Context, job config.Job) runner.Result {
	res := t.inner.Run(ctx, job)
	if len(t.tags) == 0 {
		return res
	}
	tagLine := "[tags: " + strings.Join(t.tags, ", ") + "]"
	if res.Output != "" {
		res.Output = res.Output + "\n" + tagLine
	} else {
		res.Output = tagLine
	}
	return res
}

// TagsFromJob extracts the tags slice from a job config entry.
func TagsFromJob(job config.Job) []string {
	return normaliseTags(job.Tags)
}

// WrapWithTags wraps inner with a TaggedRunner when the job defines tags;
// otherwise inner is returned unchanged.
func WrapWithTags(inner runner.Runner, job config.Job) runner.Runner {
	tags := TagsFromJob(job)
	if len(tags) == 0 {
		return inner
	}
	return NewTaggedRunner(inner, tags)
}

func normaliseTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		norm := strings.ToLower(strings.TrimSpace(t))
		if norm == "" {
			continue
		}
		if _, ok := seen[norm]; ok {
			continue
		}
		seen[norm] = struct{}{}
		out = append(out, norm)
	}
	return out
}
