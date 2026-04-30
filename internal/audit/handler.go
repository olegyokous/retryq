package audit

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const defaultLimit = 100

// Handler returns an http.HandlerFunc that streams the audit log as JSON.
// It accepts an optional ?limit=N query parameter (default 100, max log size).
func Handler(l *Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		limit := defaultLimit
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if n, err := strconv.Atoi(raw); err == nil && n > 0 {
				limit = n
			}
		}

		events := l.List()
		if limit < len(events) {
			events = events[len(events)-limit:]
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":  len(events),
			"events": events,
		})
	}
}
