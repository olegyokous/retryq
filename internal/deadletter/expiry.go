// Package deadletter provides storage and management for failed queue items
// that have exhausted all retry attempts.
package deadletter

import (
	"time"
)

// ExpiryOptions configures the expiry behaviour for dead-letter entries.
type ExpiryOptions struct {
	// MaxAge is the maximum age an entry may remain in the store before it
	// is considered expired. Zero means entries never expire.
	MaxAge time.Duration
}

// DefaultExpiryOptions returns ExpiryOptions with sensible defaults.
func DefaultExpiryOptions() ExpiryOptions {
	return ExpiryOptions{
		MaxAge: 72 * time.Hour,
	}
}

// Purge removes all entries from s whose FailedAt timestamp is older than
// maxAge. It returns the number of entries that were removed.
// If maxAge is zero, Purge is a no-op and returns 0.
func Purge(s *Store, maxAge time.Duration) int {
	if maxAge == 0 {
		return 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var kept []Entry
	removed := 0

	for _, e := range s.entries {
		if e.FailedAt.Before(cutoff) {
			removed++
		} else {
			kept = append(kept, e)
		}
	}

	s.entries = kept
	return removed
}
