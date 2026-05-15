package scheduler

import (
	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/runner"
)

// WrapWithSnapshot wraps the given runner with a SnapshotRunner when the job
// has snapshot monitoring enabled (snapshot.enabled: true in config), registers
// it with the provided SnapshotStore, and returns the wrapping runner.
// If the feature is disabled the original runner is returned unchanged.
func WrapWithSnapshot(job config.Job, r runner.Runner, store *SnapshotStore) runner.Runner {
	if !job.Snapshot.Enabled {
		return r
	}
	sr := NewSnapshotRunner(job.Name, r)
	store.Register(sr)
	return sr
}

// SnapshotConfig mirrors the snapshot block from a job's YAML definition.
type SnapshotConfig struct {
	// Enabled activates in-memory result snapshotting for the job.
	Enabled bool `yaml:"enabled"`
}
