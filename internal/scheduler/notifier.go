package scheduler

import "cronwatch/internal/runner"

// Notifier is the interface the Scheduler uses to dispatch alerts.
// It matches the concrete *notifier.Notifier so callers can pass either
// the real implementation or a test double.
type Notifier interface {
	Send(result runner.Result) error
}

// NotifierFunc is a function adapter that satisfies the Notifier interface,
// making it easy to use anonymous functions in tests.
type NotifierFunc func(result runner.Result) error

// Send implements Notifier by calling the underlying function.
func (f NotifierFunc) Send(result runner.Result) error {
	return f(result)
}
