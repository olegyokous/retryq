package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/retryq/internal/config"
	"github.com/yourorg/retryq/internal/queue"
)

func defaultConfig() config.Config {
	cfg := config.Default()
	cfg.MaxAttempts = 3
	return cfg
}

func newItem(id string) *queue.Item {
	return &queue.Item{
		ID:     id,
		Method: "POST",
		URL:    "http://example.com/hook",
		Body:   []byte(`{"event":"test"}`),
	}
}

func TestEnqueue_AddsItem(t *testing.T) {
	q := queue.New(defaultConfig())
	q.Enqueue(newItem("abc"))
	ready := q.Ready(context.Background())
	if len(ready) != 1 {
		t.Fatalf("expected 1 ready item, got %d", len(ready))
	}
	if ready[0].ID != "abc" {
		t.Errorf("expected id abc, got %s", ready[0].ID)
	}
}

func TestRecordFailure_SchedulesRetry(t *testing.T) {
	q := queue.New(defaultConfig())
	q.Enqueue(newItem("r1"))
	q.RecordFailure("r1", "timeout")
	ready := q.Ready(context.Background())
	// NextRetry is in the future, so item should not be ready immediately.
	if len(ready) != 0 {
		t.Errorf("expected 0 ready items after failure, got %d", len(ready))
	}
}

func TestRecordFailure_ExhaustsToDeadLetter(t *testing.T) {
	q := queue.New(defaultConfig())
	q.Enqueue(newItem("d1"))
	for i := 0; i < 3; i++ {
		q.RecordFailure("d1", "error")
	}
	dead := q.Dead()
	if len(dead) != 1 {
		t.Fatalf("expected 1 dead item, got %d", len(dead))
	}
	if dead[0].Status != queue.StatusDead {
		t.Errorf("expected status dead, got %s", dead[0].Status)
	}
}

func TestRecordSuccess_RemovesItem(t *testing.T) {
	q := queue.New(defaultConfig())
	q.Enqueue(newItem("s1"))
	q.RecordSuccess("s1")
	ready := q.Ready(context.Background())
	if len(ready) != 0 {
		t.Errorf("expected 0 items after success, got %d", len(ready))
	}
}

func TestIsReady_AfterNextRetry(t *testing.T) {
	item := &queue.Item{
		Status:    queue.StatusRetrying,
		NextRetry: time.Now().Add(-time.Second),
	}
	if !item.IsReady(time.Now()) {
		t.Error("expected item to be ready after NextRetry has passed")
	}
}
