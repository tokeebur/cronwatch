package runner

// Skipped is an optional field on Result indicating the run was intentionally
// not executed (e.g. due to a concurrency policy). Adding it here keeps the
// Result struct the single source of truth for run outcomes.
//
// NOTE: This file extends the Result type defined in runner.go.
// If Result is already defined there, add the Skipped field directly.
// This file documents the intended extension.
const _skippedDoc = `
Result.Skipped bool — set to true when a runner deliberately skips execution
rather than running the underlying command.  Alerters and history recorders
should treat Skipped=true runs as informational and not trigger failure alerts.
`
