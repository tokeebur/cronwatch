package scheduler

import (
	"fmt"

	"github.com/cronwatch/cronwatch/internal/config"
)

// ValidateDependencies checks that all declared dependency names in a job
// configuration refer to jobs that actually exist in the config. It returns
// an error listing any unknown dependency names.
func ValidateDependencies(cfg *config.Config) error {
	known := make(map[string]struct{}, len(cfg.Jobs))
	for _, j := range cfg.Jobs {
		known[j.Name] = struct{}{}
	}

	var errs []string
	for _, j := range cfg.Jobs {
		if raw, ok := j.Meta["depends_on"]; ok {
			deps, ok := raw.([]string)
			if !ok {
				errs = append(errs, fmt.Sprintf("job %q: depends_on must be a list of strings", j.Name))
				continue
			}
			for _, dep := range deps {
				if _, exists := known[dep]; !exists {
					errs = append(errs, fmt.Sprintf("job %q depends on unknown job %q", j.Name, dep))
				}
			}
		}
	}

	if len(errs) > 0 {
		msg := "dependency validation failed:"
		for _, e := range errs {
			msg += "\n  - " + e
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}
