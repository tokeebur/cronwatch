// Package api provides an HTTP server exposing cronwatch status endpoints.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/cronwatch/internal/history"
)

// Server is a lightweight HTTP API server.
type Server struct {
	addr    string
	store   *history.Store
	server  *http.Server
}

// New creates a new API Server bound to addr.
func New(addr string, store *history.Store) *Server {
	s := &Server{addr: addr, store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/history", s.handleHistory)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return s
}

// Start begins listening and serving HTTP requests.
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() error {
	return s.server.Close()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	all := s.store.All()
	summaries := make([]history.Summary, 0, len(all))
	for jobName := range all {
		summaries = append(summaries, history.Summarize(s.store, jobName, 24*time.Hour))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	jobName := r.URL.Query().Get("job")
	if jobName == "" {
		http.Error(w, "missing job query parameter", http.StatusBadRequest)
		return
	}
	entries, err := s.store.Last(jobName, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
