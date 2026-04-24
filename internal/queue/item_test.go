package queue_test

import (
	"testing"
	"time"

	"github.com/yourorg/retryq/internal/queue"
)

func TestIsDead_BelowMax(t *testing.T) {
	item := &queue.Item{Attempts: 2, MaxAttempts: 5}
	if item.IsDead() {
		t.Error("expected item not to be dead")
	}
}

func TestIsDead_AtMax(t *testing.T) {
	item := &queue.Item{Attempts: 5, MaxAttempts: 5}
	if !item.IsDead() {
		t.Error("expected item to be dead")
	}
}

func TestIsReady_PendingAlwaysReady(t *testing.T) {
	item := &queue.Item{Status: queue.StatusPending}
	if !item.IsReady(time.Now()) {
		t.Error("pending item should always be ready")
	}
}

func TestIsReady_RetryingBeforeNextRetry(t *testing.T) {
	item := &queue.Item{
		Status:    queue.StatusRetrying,
		NextRetry: time.Now().Add(10 * time.Second),
	}
	if item.IsReady(time.Now()) {
		t.Error("retrying item should not be ready before NextRetry")
	}
}

func TestIsReady_DeadNotReady(t *testing.T) {
	item := &queue.Item{Status: queue.StatusDead}
	if item.IsReady(time.Now()) {
		t.Error("dead item should never be ready")
	}
}

func TestIsReady_DoneNotReady(t *testing.T) {
	item := &queue.Item{Status: queue.StatusDone}
	if item.IsReady(time.Now()) {
		t.Error("done item should never be ready")
	}
}
