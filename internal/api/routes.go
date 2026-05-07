package api

import "net/http"

// registerRoutes wires all HTTP handlers to their paths with middleware applied.
func (s *Server) registerRoutes() {
	mux := s.mux

	mux.Handle("/health", chain(
		http.HandlerFunc(s.handleHealth),
		loggingMiddleware(s.logger),
		recoveryMiddleware(s.logger),
	))

	mux.Handle("/history", chain(
		http.HandlerFunc(s.handleHistory),
		loggingMiddleware(s.logger),
		recoveryMiddleware(s.logger),
	))

	mux.Handle("/status", chain(
		http.HandlerFunc(s.handleStatus),
		loggingMiddleware(s.logger),
		recoveryMiddleware(s.logger),
	))

	mux.Handle("/metrics", chain(
		http.HandlerFunc(s.handleMetrics),
		loggingMiddleware(s.logger),
		recoveryMiddleware(s.logger),
	))
}

// handleMetrics returns a plain-text Prometheus-compatible metrics snapshot.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	summaries := s.store.SummarizeAll(s.cfg.SummaryWindow)

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	for _, sm := range summaries {
		fmt.Fprintf(w, "cronwatch_runs_total{job=%q} %d\n", sm.JobName, sm.Total)
		fmt.Fprintf(w, "cronwatch_failures_total{job=%q} %d\n", sm.JobName, sm.Failures)
		fmt.Fprintf(w, "cronwatch_last_duration_ms{job=%q} %d\n", sm.JobName, sm.LastDurationMs)
	}
}
