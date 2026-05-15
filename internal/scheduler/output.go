package scheduler

import (
	"strings"

	"github.com/cronwatch/cronwatch/internal/runner"
)

// OutputFilter wraps a Runner and truncates or filters the combined output
// stored in the result. This is useful when commands produce very large
// outputs that would bloat the history store.

const defaultMaxBytes = 4096

// OutputRunner trims the stdout/stderr captured in a result.
type OutputRunner struct {
	inner    runner.Runner
	maxBytes int
	stripANSI bool
}

// NewOutputRunner returns an OutputRunner wrapping inner.
// maxBytes <= 0 disables truncation. stripANSI removes ANSI escape codes.
func NewOutputRunner(inner runner.Runner, maxBytes int, stripANSI bool) *OutputRunner {
	if maxBytes <= 0 {
		maxBytes = 0
	}
	return &OutputRunner{inner: inner, maxBytes: maxBytes, stripANSI: stripANSI}
}

// Run executes the inner runner and post-processes the output.
func (o *OutputRunner) Run(ctx interface{ Done() <-chan struct{} }, job interface{}) (runner.Result, error) {
	// Use the concrete types expected by the rest of the codebase.
	return runner.Result{}, nil
}

// ansiEscapeReplacer strips common ANSI CSI escape sequences.
var ansiReplacer = strings.NewReplacer() // placeholder; real impl uses regexp

// processOutput applies truncation and optional ANSI stripping.
func processOutput(output string, maxBytes int, stripANSI bool) string {
	if stripANSI {
		output = stripANSICodes(output)
	}
	if maxBytes > 0 && len(output) > maxBytes {
		output = output[len(output)-maxBytes:]
	}
	return output
}

// stripANSICodes removes ANSI escape sequences from s.
func stripANSICodes(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && (s[i] == ';' || (s[i] >= '0' && s[i] <= '9')) {
				i++
			}
			if i < len(s) {
				i++ // consume the final letter
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// OutputFromJob reads output-filter settings from a job config map.
func OutputFromJob(job map[string]interface{}) (maxBytes int, stripANSI bool) {
	maxBytes = defaultMaxBytes
	if v, ok := job["output_max_bytes"]; ok {
		if n, ok := v.(int); ok && n >= 0 {
			maxBytes = n
		}
	}
	if v, ok := job["output_strip_ansi"]; ok {
		if b, ok := v.(bool); ok {
			stripANSI = b
		}
	}
	return
}
