package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/runner"
)

// TestCircuitBreakerFromJob_Defaults verifies that missing fields fall back to defaults.
func TestCircuitBreakerFromJob_Defaults(t *testing.T) {
	job := config.Job{
		CircuitBreaker: &config.CBConfig{},
	}
	p := CircuitBreakerFromJob(job)
	if !p.Enabled {
		t.Fatal("expected enabled")
	}
	if p.MaxFailures != defaultCBMaxFailures {
		t.Fatalf("expected %d, got %d", defaultCBMaxFailures, p.MaxFailures)
	}
	if p.Cooldown != defaultCBCooldown {
		t.Fatalf("expected %v, got %v", defaultCBCooldown, p.Cooldown)
	}
}

// TestCircuitBreakerFromJob_Disabled verifies nil config disables the breaker.
func TestCircuitBreakerFromJob_Disabled(t *testing.T) {
	p := CircuitBreakerFromJob(config.Job{})
	if p.Enabled {
		t.Fatal("expected disabled when no circuit_breaker config")
	}
}

// TestCircuitBreakerFromJob_CustomValues verifies custom values are respected.
func TestCircuitBreakerFromJob_CustomValues(t *testing.T) {
	job := config.Job{
		CircuitBreaker: &config.CBConfig{
			MaxFailures:     2,
			CooldownSeconds: 30,
		},
	}
	p := CircuitBreakerFromJob(job)
	if p.MaxFailures != 2 {
		t.Fatalf("expected 2, got %d", p.MaxFailures)
	}
	if p.Cooldown != 30*time.Second {
		t.Fatalf("expected 30s, got %v", p.Cooldown)
	}
}

// TestCircuitBreaker_ResetOnSuccess verifies failure counter resets after a success.
func TestCircuitBreaker_ResetOnSuccess(t *testing.T) {
	boomed := errors.New("boom")
	sr := &stubRunner{
		results: []runner.Result{
			failResult(boomed),
			okResult(),
			failResult(boomed),
		},
	}
	cb := NewCircuitBreaker(sr, 2, time.Minute)
	cb.Run(context.Background(), runner.Job{Name: "j"}) // 1 failure
	cb.Run(context.Background(), runner.Job{Name: "j"}) // success → reset
	cb.Run(context.Background(), runner.Job{Name: "j"}) // 1 failure again
	if cb.State() != StateClosed {
		t.Fatalf("expected closed (only 1 failure after reset), got %v", cb.State())
	}
}
