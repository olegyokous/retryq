package server

import (
	"encoding/json"
	"net/http"

	"github.com/yourorg/retryq/internal/deadletter"
)

// handleDeadLetter returns all dead-letter entries as a JSON array.
// It only accepts GET requests and responds with 405 for anything else.
func handleDeadLetter(store *deadletter.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		entries := store.List()

		// Return an empty array rather than null when there are no entries.
		if entries == nil {
			entries = []*deadletter.Entry{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(entries); err != nil {
			// Headers already sent; log the error via the response trailer.
			w.Header().Set("X-Encode-Error", err.Error())
		}
	}
}
