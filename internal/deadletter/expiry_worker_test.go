package deadletter_test

import (
	"testing"
	"time"

	"github.com/andres-movl/retryq/internal/deadletter"
)

func makeExpiredEntryForWorker(id string, age time.Duration) deadletter.Entry {
	return deadletter.Entry{
		ID:        id,
		TargetURL: "http://example.com",
		Payload:   []byte(`{"key":"value"}`),
		FailedAt:  time.Now().Add(-age),
		Attempts:  3,
	}
}

func TestNewExpiryWorker_StartsAndStops(t *testing.T) {
	store := New(10)

	opts := deadletter.DefaultExpiryOptions()
	opts.Interval = 20 * time.Millisecond
	opts.MaxAge = 50 * time.Millisecond

	worker := deadletter.NewExpiryWorker(store, opts)

	worker.Start()
	time.Sleep(30 * time.Millisecond)
	worker.Stop()
	// Should not panic or deadlock
}

func TestExpiryWorker_PurgesExpiredEntries(t *testing.T) {
	store := New(10)

	// Add an entry that is already expired
	expired := makeExpiredEntryForWorker("expired-1", 200*time.Millisecond)
	store.Add(expired)

	// Add a fresh entry
	fresh := deadletter.Entry{
		ID:        "fresh-1",
		TargetURL: "http://example.com",
		Payload:   []byte(`{}`),
		FailedAt:  time.Now(),
		Attempts:  1,
	}
	store.Add(fresh)

	opts := deadletter.ExpiryOptions{
		MaxAge:   100 * time.Millisecond,
		Interval: 20 * time.Millisecond,
	}

	worker := deadletter.NewExpiryWorker(store, opts)
	worker.Start()
	defer worker.Stop()

	// Wait for at least one purge cycle
	time.Sleep(50 * time.Millisecond)

	entries := store.List()
	for _, e := range entries {
		if e.ID == "expired-1" {
			t.Errorf("expected expired entry to be purged, but it still exists")
		}
	}

	// Fresh entry should remain
	found := false
	for _, e := range entries {
		if e.ID == "fresh-1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected fresh entry to remain after purge")
	}
}

func TestExpiryWorker_StopIsIdempotent(t *testing.T) {
	store := New(10)
	opts := deadletter.DefaultExpiryOptions()
	opts.Interval = 10 * time.Millisecond

	worker := deadletter.NewExpiryWorker(store, opts)
	worker.Start()
	worker.Stop()
	// Calling Stop again should not panic
	worker.Stop()
}

func TestExpiryWorker_ZeroMaxAge_DoesNotPurge(t *testing.T) {
	store := New(10)

	entry := makeExpiredEntryForWorker("entry-1", 1*time.Hour)
	store.Add(entry)

	opts := deadletter.ExpiryOptions{
		MaxAge:   0, // disabled
		Interval: 15 * time.Millisecond,
	}

	worker := deadletter.NewExpiryWorker(store, opts)
	worker.Start()
	defer worker.Stop()

	time.Sleep(40 * time.Millisecond)

	entries := store.List()
	if len(entries) != 1 {
		t.Errorf("expected 1 entry to remain when MaxAge=0, got %d", len(entries))
	}
}
