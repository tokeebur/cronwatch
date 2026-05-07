// Package history provides a lightweight, file-backed store for recording
// cron job execution results. Entries are persisted as a JSON array so they
// survive process restarts and can be inspected with standard tooling.
//
// Typical usage:
//
//	store, err := history.New("/var/lib/cronwatch/history.json")
//	if err != nil { ... }
//
//	store.Record(history.Entry{
//	    JobName:  "daily-backup",
//	    ExitCode: result.ExitCode,
//	    Duration: result.Duration,
//	})
package history
