package deadletter

import (
	"encoding/json"
	"net/http"
)

// RequeueRequest is the expected JSON body for a requeue request.
type RequeueRequest struct {
	ID string `json:"id"`
}

// RequeueResponse is returned after a successful requeue.
type RequeueResponse struct {
	ID      string `json:"id"`
	Requeued bool   `json:"requeued"`
}

// HandleRequeue handles POST /dead-letter/requeue requests.
// It pops the entry from the dead-letter store and re-enqueues it
// via the provided enqueue function.
func HandleRequeue(s *Store, enqueue func(Entry) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RequeueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
			http.Error(w, "invalid request body: 'id' is required", http.StatusBadRequest)
			return
		}

		entry, ok := s.Pop(req.ID)
		if !ok {
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}

		if err := enqueue(entry); err != nil {
			// Put it back to avoid losing the entry on enqueue failure.
			s.Add(entry)
			http.Error(w, "failed to re-enqueue entry", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(RequeueResponse{ID: entry.ID, Requeued: true})
	}
}
