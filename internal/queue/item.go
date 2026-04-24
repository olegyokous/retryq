package queue

import (
	"net/http"
	"time"
)

// Status represents the current state of a queue item.
type Status string

const (
	StatusPending  Status = "pending"
	StatusRetrying Status = "retrying"
	StatusDead     Status = "dead"
	StatusDone     Status = "done"
)

// Item represents a single HTTP request to be retried.
type Item struct {
	ID          string
	Method      string
	URL         string
	Headers     http.Header
	Body        []byte
	Attempts    int
	MaxAttempts int
	NextRetry   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Status      Status
	LastError   string
}

// IsDead returns true when the item has exhausted all retry attempts.
func (i *Item) IsDead() bool {
	return i.Attempts >= i.MaxAttempts
}

// IsReady returns true when the item is eligible to be processed.
func (i *Item) IsReady(now time.Time) bool {
	return i.Status == StatusPending || (i.Status == StatusRetrying && !now.Before(i.NextRetry))
}
