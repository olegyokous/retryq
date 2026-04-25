package server

import (
	"encoding/json"
	"net/http"
	"time"
)

// metricsSnapshot is the JSON shape returned by GET /metrics/snapshot.
type metricsSnapshot struct {
	Timestamp   time.Time        `json:"timestamp"`
	QueueDepth  map[string]int64 `json:"queue_depth"`
	Enqueued    int64            `json:"enqueued_total"`
	Retried     int64            `json:"retried_total"`
	DeadLetters int64            `json:"dead_letters_total"`
	Dispatched  int64            `json:"dispatched_total"`
}

// handleMetricsSnapshot serves a lightweight JSON summary of the current
// queue metrics. It is intentionally separate from the Prometheus /metrics
// endpoint so that operators can poll a human-readable snapshot without
// needing a full Prometheus stack.
func (s *Server) handleMetricsSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snap := metricsSnapshot{
		Timestamp: time.Now().UTC(),
		QueueDepth: map[string]int64{
			"pending":    s.metrics.QueueDepthValue("pending"),
			"retrying":   s.metrics.QueueDepthValue("retrying"),
			"dead":       s.metrics.QueueDepthValue("dead"),
		},
		Enqueued:    s.metrics.EnqueuedTotal(),
		Retried:     s.metrics.RetriedTotal(),
		DeadLetters: s.metrics.DeadLettersTotal(),
		Dispatched:  s.metrics.DispatchedTotal(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(snap)
}
