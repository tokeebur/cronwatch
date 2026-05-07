// Package history records the execution history of cron jobs managed by
// cronwatch. Each job run is stored as an Entry containing timing information,
// exit status, and captured output.
//
// The Store type persists entries to a JSON file on disk so that history
// survives daemon restarts. Use New to create a Store backed by a given
// directory.
//
// The Cleaner type provides a Prune method that removes entries older than a
// configurable retention duration, keeping the on-disk store from growing
// without bound.
package history
