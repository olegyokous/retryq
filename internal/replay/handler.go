package replay

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/retryq/internal/deadletter"
)

// storeDispatcher bridges the deadletter.Store and the Dispatcher interface
// so the HTTP handler can stay decoupled from concrete types.
type storeDispatcher interface {
	Dispatcher
}

// Store is the subset of deadletter.Store used by the handler.
type Store interface {
	List() []deadletter.Entry
}

// replayResponse is the JSON body returned after a bulk replay.
type replayResponse struct {
	Total    int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

// NewHandler returns an http.Handler that replays all current dead-letter
// entries through the provided Dispatcher.
func NewHandler(store Store, d Dispatcher, opts Options) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		entries := store.List()
		results := Replay(r.Context(), entries, d, opts)

		var succeeded, failed int
		for _, res := range results {
			if res.Err == nil {
				succeeded++
			} else {
				failed++
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(replayResponse{
			Total:     len(results),
			Succeeded: succeeded,
			Failed:    failed,
		})
	})
}
