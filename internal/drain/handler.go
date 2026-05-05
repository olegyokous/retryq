package drain

import (
	"encoding/json"
	"net/http"
)

// StatusResponse is the JSON body returned by the drain status endpoint.
type StatusResponse struct {
	Inflight int  `json:"inflight"`
	Drained  bool `json:"drained"`
}

// NewHandler returns an http.Handler that reports the current drain
// status.  Callers can poll this endpoint to determine when it is safe
// to terminate the process.
//
//	GET /drain  →  200 {"inflight":0,"drained":true}
func NewHandler(w Waiter) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		n := w.Inflight()
		resp := StatusResponse{
			Inflight: n,
			Drained:  n == 0,
		}
		rw.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(rw).Encode(resp)
	})
}
