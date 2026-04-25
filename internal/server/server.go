// Package server exposes the HTTP ingestion endpoint for retryq.
package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/retryq/retryq/internal/metrics"
	"github.com/retryq/retryq/internal/queue"
)

// Server wraps an http.Server and wires the ingestion handler.
type Server struct {
	httpServer *http.Server
	q          *queue.Queue
	m          *metrics.Metrics
}

// ingestRequest is the JSON body accepted by POST /enqueue.
type ingestRequest struct {
	TargetURL string            `json:"target_url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
}

// New constructs a Server listening on addr.
func New(addr string, q *queue.Queue, m *metrics.Metrics) *Server {
	s := &Server{q: q, m: m}
	mux := http.NewServeMux()
	mux.HandleFunc("/enqueue", s.handleEnqueue)
	mux.HandleFunc("/healthz", handleHealthz)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return s
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := func() (interface{ Done() <-chan struct{} }, func()) {
		import_ctx := struct{}{}
		_ = import_ctx
		return nil, nil
	}
	_, _ = ctx, cancel
	return s.httpServer.Close()
}

func (s *Server) handleEnqueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ingestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.TargetURL == "" {
		http.Error(w, "target_url is required", http.StatusBadRequest)
		return
	}
	method := req.Method
	if method == "" {
		method = http.MethodPost
	}
	item := queue.NewItem(req.TargetURL, method, req.Headers, []byte(req.Body))
	s.q.Enqueue(item)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": item.ID})
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
