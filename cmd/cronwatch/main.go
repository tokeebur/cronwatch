// Package main is the entry point for the cronwatch daemon.
// It wires together configuration, scheduling, alerting, history,
// and the HTTP API into a single long-running process.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/cronwatch/internal/api"
	"github.com/example/cronwatch/internal/config"
	"github.com/example/cronwatch/internal/history"
	"github.com/example/cronwatch/internal/notifier"
	"github.com/example/cronwatch/internal/runner"
	"github.com/example/cronwatch/internal/scheduler"
	"github.com/example/cronwatch/internal/webhook"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "cronwatch: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfgPath := flag.String("config", "configs/cronwatch.yaml", "path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// History store — persists run records across restarts.
	store, err := history.New(cfg.HistoryPath)
	if err != nil {
		return fmt.Errorf("open history store: %w", err)
	}

	// Alerters — build the chain: email + webhook, both throttled.
	var alerters []scheduler.Alerter

	if cfg.SMTP.Host != "" {
		n, nErr := notifier.New(cfg.SMTP)
		if nErr != nil {
			return fmt.Errorf("create notifier: %w", nErr)
		}
		alerters = append(alerters, scheduler.NewThrottledAlerter(n, scheduler.DefaultThrottle))
	}

	if cfg.Webhook.URL != "" {
		w := webhook.New(cfg.Webhook)
		alerters = append(alerters, scheduler.NewThrottledAlerter(
			scheduler.NewWebhookAlerter(w),
			scheduler.DefaultThrottle,
		))
	}

	multi := scheduler.NewMultiAlerter(alerters...)

	// Scheduler — wraps the runner with retry, circuit-breaker, and history.
	r := runner.New()
	sched := scheduler.New(cfg, r, multi, store)

	// HTTP API — exposes /health, /history, /status, /metrics.
	apiCfg := api.DefaultConfig()
	if cfg.API.Addr != "" {
		apiCfg.Addr = cfg.API.Addr
	}
	srv := api.New(apiCfg, store)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start the API server in the background.
	httpServer := &http.Server{
		Addr:    apiCfg.Addr,
		Handler: srv,
	}
	go func() {
		slog.Info("api listening", "addr", apiCfg.Addr)
		if sErr := httpServer.ListenAndServe(); sErr != nil && sErr != http.ErrServerClosed {
			slog.Error("api server error", "err", sErr)
		}
	}()

	// Start the scheduler — blocks until context is cancelled.
	slog.Info("cronwatch started", "jobs", len(cfg.Jobs))
	sched.Start(ctx)

	// Graceful shutdown.
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if sErr := httpServer.Shutdown(shutCtx); sErr != nil {
		slog.Warn("api shutdown error", "err", sErr)
	}

	slog.Info("cronwatch stopped")
	return nil
}
