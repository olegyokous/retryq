// Package audit provides a structured event log for retry queue operations,
// recording enqueue, dispatch, retry, and dead-letter events for observability.
package audit

import (
	"sync"
	"time"
)

// EventKind classifies an audit event.
type EventKind string

const (
	EventEnqueued   EventKind = "enqueued"
	EventDispatched EventKind = "dispatched"
	EventRetry      EventKind = "retry"
	EventDeadLetter EventKind = "dead_letter"
	EventRequeued   EventKind = "requeued"
)

// Event represents a single auditable action in the retry pipeline.
type Event struct {
	ID        string            `json:"id"`
	Kind      EventKind         `json:"kind"`
	TargetURL string            `json:"target_url"`
	Attempt   int               `json:"attempt"`
	Status    int               `json:"status,omitempty"`
	Error     string            `json:"error,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// Log is a bounded, thread-safe in-memory audit log.
type Log struct {
	mu      sync.RWMutex
	events  []Event
	maxSize int
}

// DefaultMaxSize is the default capacity of the audit log.
const DefaultMaxSize = 1000

// New returns a new Log with the given maximum size.
// If maxSize <= 0, DefaultMaxSize is used.
func New(maxSize int) *Log {
	if maxSize <= 0 {
		maxSize = DefaultMaxSize
	}
	return &Log{
		events:  make([]Event, 0, maxSize),
		maxSize: maxSize,
	}
}

// Record appends an event to the log, evicting the oldest entry when full.
func (l *Log) Record(e Event) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.events) >= l.maxSize {
		l.events = l.events[1:]
	}
	l.events = append(l.events, e)
}

// List returns a shallow copy of all recorded events.
func (l *Log) List() []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}

// Len returns the current number of recorded events.
func (l *Log) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.events)
}
