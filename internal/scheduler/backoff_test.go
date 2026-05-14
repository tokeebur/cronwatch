package scheduler

import (
	"testing"
	"time"
)

func TestBackoffPolicy_Fixed(t *testing.T) {
	p := BackoffPolicy{
		Strategy:  BackoffFixed,
		BaseDelay: 2 * time.Second,
		MaxDelay:  0,
	}
	for attempt := 0; attempt < 5; attempt++ {
		if got := p.Delay(attempt); got != 2*time.Second {
			t.Errorf("attempt %d: want 2s, got %v", attempt, got)
		}
	}
}

func TestBackoffPolicy_Linear(t *testing.T) {
	p := BackoffPolicy{
		Strategy:  BackoffLinear,
		BaseDelay: 1 * time.Second,
		MaxDelay:  0,
	}
	want := []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second}
	for i, w := range want {
		if got := p.Delay(i); got != w {
			t.Errorf("attempt %d: want %v, got %v", i, w, got)
		}
	}
}

func TestBackoffPolicy_Exponential(t *testing.T) {
	p := BackoffPolicy{
		Strategy:  BackoffExponential,
		BaseDelay: 1 * time.Second,
		MaxDelay:  0,
	}
	want := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}
	for i, w := range want {
		if got := p.Delay(i); got != w {
			t.Errorf("attempt %d: want %v, got %v", i, w, got)
		}
	}
}

func TestBackoffPolicy_MaxDelayCaps(t *testing.T) {
	p := BackoffPolicy{
		Strategy:  BackoffExponential,
		BaseDelay: 1 * time.Second,
		MaxDelay:  5 * time.Second,
	}
	for attempt := 3; attempt < 10; attempt++ {
		if got := p.Delay(attempt); got > 5*time.Second {
			t.Errorf("attempt %d: expected delay capped at 5s, got %v", attempt, got)
		}
	}
}

func TestBackoffPolicy_NegativeAttemptTreatedAsZero(t *testing.T) {
	p := DefaultBackoffPolicy()
	got := p.Delay(-1)
	expected := p.Delay(0)
	if got != expected {
		t.Errorf("negative attempt: want %v, got %v", expected, got)
	}
}

// TestBackoffPolicy_UnknownStrategyFallback verifies that an unrecognized
// strategy falls back to the fixed delay behavior rather than panicking.
func TestBackoffPolicy_UnknownStrategyFallback(t *testing.T) {
	p := BackoffPolicy{
		Strategy:  "unknown",
		BaseDelay: 3 * time.Second,
		MaxDelay:  0,
	}
	for attempt := 0; attempt < 3; attempt++ {
		if got := p.Delay(attempt); got != 3*time.Second {
			t.Errorf("attempt %d: unknown strategy should fall back to fixed, got %v", attempt, got)
		}
	}
}

func TestBackoffFromJob_Defaults(t *testing.T) {
	p := BackoffFromJob(map[string]string{})
	def := DefaultBackoffPolicy()
	if p.Strategy != def.Strategy || p.BaseDelay != def.BaseDelay || p.MaxDelay != def.MaxDelay {
		t.Errorf("expected default policy, got %+v", p)
	}
}

func TestBackoffFromJob_ParsesAllFields(t *testing.T) {
	opts := map[string]string{
		"backoff_strategy": "linear",
		"backoff_base":     "500ms",
		"backoff_max":      "10s",
	}
	p := BackoffFromJob(opts)
	if p.Strategy != BackoffLinear {
		t.Errorf("want linear strategy, got %v", p.Strategy)
	}
	if p.BaseDelay != 500*time.Millisecond {
		t.Errorf("want 500ms base, got %v", p.BaseDelay)
	}
	if p.MaxDelay != 10*time.Second {
		t.Errorf("want 10s max, got %v", p.MaxDelay)
	}
}

func TestBackoffFromJob_InvalidDurationIgnored(t *testing.T) {
	opts := map[string]string{
		"backoff_base": "not-a-duration",
		"backoff_max":  "also-bad",
	}
	p := BackoffFromJob(opts)
	def := DefaultBackoffPolicy()
	if p.BaseDelay != def.BaseDelay {
		t.Errorf("bad base should be ignored, got %v", p.BaseDelay)
	}
	if p.MaxDelay != def.MaxDelay {
		t.Errorf("bad max should be ignored, got %v", p.MaxDelay)
	}
}
