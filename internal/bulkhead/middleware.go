package bulkhead

import (
	"net/http"
	"strconv"
)

// Middleware returns an http.Handler that enforces the bulkhead limit.
// When the bulkhead is full it responds with 503 Service Unavailable and
// sets the X-Bulkhead-Limit header so clients can introspect the cap.
func NewMiddleware(b *Bulkhead, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if b != nil {
			w.Header().Set("X-Bulkhead-Limit", strconv.FormatInt(b.max, 10))
			if err := b.Acquire(); err != nil {
				w.Header().Set("X-Bulkhead-Inflight", strconv.FormatInt(b.Inflight(), 10))
				http.Error(w, "service at capacity", http.StatusServiceUnavailable)
				return
			}
			defer b.Release()
		}
		next.ServeHTTP(w, r)
	})
}
