// Package dedupe provides idempotency for enqueued retry items by tracking
// recently seen request IDs and rejecting duplicates within a configurable TTL.
package dedupe

import (
	"sync"
	"time"
)

// Options configures the deduplication store.
type Options struct {
	// TTL is how long an ID is remembered after it was first seen.
	// Defaults to 5 minutes.
	TTL time.Duration

	// MaxSize is the maximum number of IDs held in memory.
	// Oldest entries are evicted when the limit is reached.
	// Zero means no limit.
	MaxSize int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		TTL:     5 * time.Minute,
		MaxSize: 10_000,
	}
}

type entry struct {
	id        string
	expiresAt time.Time
}

// Store tracks seen IDs and reports duplicates.
type Store struct {
	mu      sync.Mutex
	opts    Options
	seen    map[string]time.Time
	ordered []entry // insertion-order for eviction
}

// New creates a Store with the given options.
func New(opts Options) *Store {
	if opts.TTL <= 0 {
		opts.TTL = DefaultOptions().TTL
	}
	return &Store{
		opts: opts,
		seen: make(map[string]time.Time),
	}
}

// IsDuplicate returns true if id was seen within the TTL window.
// If id is new it is recorded and false is returned.
func (s *Store) IsDuplicate(id string) bool {
	if id == "" {
		return false
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	s.evictExpired(now)

	if exp, ok := s.seen[id]; ok && now.Before(exp) {
		return true
	}

	// Evict oldest when at capacity.
	if s.opts.MaxSize > 0 && len(s.seen) >= s.opts.MaxSize {
		s.evictOldest()
	}

	exp := now.Add(s.opts.TTL)
	s.seen[id] = exp
	s.ordered = append(s.ordered, entry{id: id, expiresAt: exp})
	return false
}

// Len returns the number of currently tracked IDs.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.seen)
}

func (s *Store) evictExpired(now time.Time) {
	i := 0
	for i < len(s.ordered) && now.After(s.ordered[i].expiresAt) {
		delete(s.seen, s.ordered[i].id)
		i++
	}
	s.ordered = s.ordered[i:]
}

func (s *Store) evictOldest() {
	if len(s.ordered) == 0 {
		return
	}
	delete(s.seen, s.ordered[0].id)
	s.ordered = s.ordered[1:]
}
