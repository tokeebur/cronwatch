package scheduler

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cronwatch/cronwatch/internal/config"
	"github.com/cronwatch/cronwatch/internal/runner"
)

// EnvRunner wraps a Runner and injects extra environment variables into each
// job execution. Variables are merged with the current process environment;
// values defined in the job config take precedence over inherited ones.
type EnvRunner struct {
	inner runner.Runner
	vars  []string // KEY=VALUE pairs
}

// NewEnvRunner returns an EnvRunner that prepends extraVars (KEY=VALUE) to the
// environment forwarded to the inner runner.
func NewEnvRunner(inner runner.Runner, extraVars []string) *EnvRunner {
	return &EnvRunner{inner: inner, vars: extraVars}
}

// Run merges the configured environment variables into ctx before delegating
// to the wrapped runner.
func (e *EnvRunner) Run(ctx context.Context, job config.Job) (runner.Result, error) {
	ctx = contextWithEnv(ctx, e.vars)
	return e.inner.Run(ctx, job)
}

// contextKey is an unexported type for context keys in this package.
type contextKey int

const envKey contextKey = iota

// contextWithEnv stores merged environment variables in ctx.
func contextWithEnv(ctx context.Context, extra []string) context.Context {
	merged := mergeEnv(os.Environ(), extra)
	return context.WithValue(ctx, envKey, merged)
}

// EnvFromContext retrieves the merged environment slice stored by EnvRunner.
// Returns nil if no environment was injected.
func EnvFromContext(ctx context.Context) []string {
	v, _ := ctx.Value(envKey).([]string)
	return v
}

// mergeEnv combines base with overrides. Keys present in overrides replace
// those in base. Order of base is preserved; new keys from overrides are
// appended.
func mergeEnv(base, overrides []string) []string {
	result := make([]string, 0, len(base)+len(overrides))
	index := make(map[string]int, len(base))

	for _, kv := range base {
		idx := len(result)
		result = append(result, kv)
		if k := envKey_(kv); k != "" {
			index[k] = idx
		}
	}

	for _, kv := range overrides {
		k := envKey_(kv)
		if k == "" {
			continue
		}
		if idx, ok := index[k]; ok {
			result[idx] = kv
		} else {
			result = append(result, kv)
		}
	}
	return result
}

// envKey_ returns the key portion of a KEY=VALUE string.
func envKey_(kv string) string {
	if i := strings.IndexByte(kv, '='); i > 0 {
		return kv[:i]
	}
	return ""
}

// EnvFromJob converts the job's Env map into KEY=VALUE pairs.
func EnvFromJob(job config.Job) []string {
	pairs := make([]string, 0, len(job.Env))
	for k, v := range job.Env {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return pairs
}

// WrapWithEnv wraps inner with an EnvRunner when the job defines environment
// variables; otherwise it returns inner unchanged.
func WrapWithEnv(inner runner.Runner, job config.Job) runner.Runner {
	if len(job.Env) == 0 {
		return inner
	}
	return NewEnvRunner(inner, EnvFromJob(job))
}
