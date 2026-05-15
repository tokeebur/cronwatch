package scheduler

import (
	"testing"

	"github.com/cronwatch/cronwatch/internal/config"
)

func jobWithTimeoutAlertOptions(opts map[string]interface{}) config.Job {
	return config.Job{
		Name:    "test-job",
		Command: "echo hi",
		Options: opts,
	}
}

func TestTimeoutAlertFromJob_Defaults(t *testing.T) {
	p := TimeoutAlertFromJob(config.Job{Name: "j", Command: "echo"})
	if !p.Enabled {
		t.Error("expected Enabled=true by default")
	}
	if !p.Forward {
		t.Error("expected Forward=true by default")
	}
}

func TestTimeoutAlertFromJob_DisabledViaOptions(t *testing.T) {
	job := jobWithTimeoutAlertOptions(map[string]interface{}{
		"timeout_alert_enabled": false,
	})
	p := TimeoutAlertFromJob(job)
	if p.Enabled {
		t.Error("expected Enabled=false")
	}
	if !p.Forward {
		t.Error("Forward should remain true when not explicitly set")
	}
}

func TestTimeoutAlertFromJob_ForwardFalse(t *testing.T) {
	job := jobWithTimeoutAlertOptions(map[string]interface{}{
		"timeout_alert_forward": false,
	})
	p := TimeoutAlertFromJob(job)
	if !p.Enabled {
		t.Error("expected Enabled=true")
	}
	if p.Forward {
		t.Error("expected Forward=false")
	}
}

func TestWrapWithTimeoutAlert_ReturnsTimeoutAlerterByDefault(t *testing.T) {
	cap := &captureAlerter{}
	job := config.Job{Name: "j", Command: "echo"}
	wrapped := WrapWithTimeoutAlert(cap, job)
	if _, ok := wrapped.(*TimeoutAlerter); !ok {
		t.Fatal("expected *TimeoutAlerter when enabled by default")
	}
}
