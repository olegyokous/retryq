package server

import (
	"net/http"
)

// ServeHTTP makes Server implement http.Handler so it can be used
// directly in tests via httptest.NewRecorder without starting a real
// TCP listener.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpServer.Handler.ServeHTTP(w, r)
}
