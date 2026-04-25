// Package deadletter provides storage and requeue support for exhausted items.
package deadletter

import (
	"errors"
	"time"
)

// ErrNotFound is returned when an entry with the given ID does not exist.
var ErrNotFound = errors.New("deadletter: entry not found")

// RequeuedEntry wraps an Entry with the timestamp it was requeued at.
type RequeuedEntry struct {
	Entry
	RequeuedAt time.Time
}

// Pop removes and returns the entry with the given ID from the dead-letter
// store. Returns ErrNotFound if no such entry exists.
func (s *Store) Pop(id string) (Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, e := range s.entries {
		if e.ID == id {
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			return e, nil
		}
	}
	return Entry{}, ErrNotFound
}

// Len returns the current number of entries in the store.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}
