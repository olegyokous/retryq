package queue

import (
	"context"
	"sync"
	"time"

	"github.com/yourorg/retryq/internal/backoff"
	"github.com/yourorg/retryq/internal/config"
)

// Queue manages pending retry items in memory.
type Queue struct {
	mu     sync.Mutex
	items  map[string]*Item
	policy *backoff.Policy
	cfg    config.Config
}

// New creates a new Queue with the provided configuration.
func New(cfg config.Config) *Queue {
	return &Queue{
		items:  make(map[string]*Item),
		policy: backoff.New(cfg),
		cfg:    cfg,
	}
}

// Enqueue adds a new item to the queue.
func (q *Queue) Enqueue(item *Item) {
	q.mu.Lock()
	defer q.mu.Unlock()
	item.Status = StatusPending
	item.CreatedAt = time.Now()
	item.UpdatedAt = item.CreatedAt
	item.MaxAttempts = q.cfg.MaxAttempts
	q.items[item.ID] = item
}

// Ready returns all items that are ready to be processed.
func (q *Queue) Ready(ctx context.Context) []*Item {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now()
	var ready []*Item
	for _, item := range q.items {
		if item.IsReady(now) {
			ready = append(ready, item)
		}
	}
	return ready
}

// RecordFailure marks an item as failed and schedules the next retry.
func (q *Queue) RecordFailure(id, errMsg string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	item, ok := q.items[id]
	if !ok {
		return
	}
	item.Attempts++
	item.LastError = errMsg
	item.UpdatedAt = time.Now()
	if item.IsDead() {
		item.Status = StatusDead
		return
	}
	item.Status = StatusRetrying
	item.NextRetry = time.Now().Add(q.policy.Next(item.Attempts))
}

// RecordSuccess marks an item as successfully delivered and removes it.
func (q *Queue) RecordSuccess(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.items, id)
}

// Dead returns all dead-letter items.
func (q *Queue) Dead() []*Item {
	q.mu.Lock()
	defer q.mu.Unlock()
	var dead []*Item
	for _, item := range q.items {
		if item.Status == StatusDead {
			dead = append(dead, item)
		}
	}
	return dead
}
