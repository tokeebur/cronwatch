// Package runner provides functionality for executing shell commands
// on behalf of monitored cron jobs.
//
// A Runner wraps exec.Cmd with optional timeout support and captures
// exit codes, combined output, and wall-clock duration into a Result
// value that downstream alerting logic can inspect.
package runner
