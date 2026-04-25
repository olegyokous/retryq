// Package deadletter provides storage and retrieval for dead-letter queue items
// that have exhausted all retry attempts.
package deadletter

import (
	"sync"
	"time"
)

// Entry represents a dead-lettered item with metadata about why it failed.
type Entry struct {
	ID          string            `json:"id"`
	TargetURL   string            `json:"target_url"`
	Payload     []byte            `json:"payload"`
	Headers     map[string]string `json:"headers"`
	Attempts    int               `json:"attempts"`
	LastError   string            `json:"last_error"`
	DeadAt      time.Time         `json:"dead_at"`
}

// Store holds dead-letter entries in memory.
type Store struct {
	mu      sync.RWMutex
	entries []*Entry
	maxSize int
}

// New creates a new Store with the given maximum capacity.
// When the store is full, the oldest entry is evicted.
func New(maxSize int) *Store {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &Store{
		entries: make([]*Entry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add appends a new dead-letter entry to the store.
// If the store is at capacity the oldest entry is dropped.
func (s *Store) Add(e *Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.entries) >= s.maxSize {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, e)
}

// List returns a shallow copy of all dead-letter entries.
func (s *Store) List() []*Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*Entry, len(s.entries))
	copy(out, s.entries)
	return out
}

// Len returns the current number of entries in the store.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
