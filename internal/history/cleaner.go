package history

import (
	"time"
)

// Cleaner removes old history entries beyond a retention period.
type Cleaner struct {
	store     *Store
	retention time.Duration
}

// NewCleaner creates a Cleaner that will remove entries older than retention.
func NewCleaner(store *Store, retention time.Duration) *Cleaner {
	return &Cleaner{
		store:     store,
		retention: retention,
	}
}

// Prune removes all history entries whose FinishedAt time is older than the
// configured retention duration. It returns the number of entries removed.
func (c *Cleaner) Prune() (int, error) {
	c.store.mu.Lock()
	defer c.store.mu.Unlock()

	cutoff := time.Now().Add(-c.retention)
	removed := 0

	for job, entries := range c.store.data {
		var kept []Entry
		for _, e := range entries {
			if e.FinishedAt.After(cutoff) {
				kept = append(kept, e)
			} else {
				removed++
			}
		}
		if len(kept) == 0 {
			delete(c.store.data, job)
		} else {
			c.store.data[job] = kept
		}
	}

	if err := c.store.persist(); err != nil {
		return removed, err
	}
	return removed, nil
}
