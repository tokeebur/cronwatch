package api

import "net/http"

// ServeHTTP allows *Server to satisfy http.Handler for use in tests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.Handler.ServeHTTP(w, r)
}
