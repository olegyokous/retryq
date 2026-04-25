package deadletter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/yourorg/retryq/internal/deadletter"
)

func newEntry(id string) *deadletter.Entry {
	return &deadletter.Entry{
		ID:        id,
		TargetURL: "http://example.com",
		Payload:   []byte(`{"hello":"world"}`),
		Attempts:  3,
		LastError: "connection refused",
		DeadAt:    time.Now(),
	}
}

func TestAdd_StoresEntry(t *testing.T) {
	s := deadletter.New(10)
	s.Add(newEntry("a"))

	if got := s.Len(); got != 1 {
		t.Fatalf("expected 1 entry, got %d", got)
	}
}

func TestList_ReturnsCopy(t *testing.T) {
	s := deadletter.New(10)
	s.Add(newEntry("a"))
	s.Add(newEntry("b"))

	list := s.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}

	// Mutating the returned slice must not affect the store.
	list[0] = nil
	if s.Len() != 2 {
		t.Fatal("store was mutated via returned slice")
	}
}

func TestAdd_EvictsOldestWhenFull(t *testing.T) {
	const max = 5
	s := deadletter.New(max)

	for i := 0; i < max+3; i++ {
		s.Add(newEntry(fmt.Sprintf("id-%d", i)))
	}

	if got := s.Len(); got != max {
		t.Fatalf("expected store capped at %d, got %d", max, got)
	}

	// The first entry should now be id-3 (oldest three evicted).
	list := s.List()
	if list[0].ID != "id-3" {
		t.Fatalf("expected oldest surviving entry id-3, got %s", list[0].ID)
	}
}

func TestNew_DefaultMaxSize(t *testing.T) {
	s := deadletter.New(0) // zero triggers default
	if s == nil {
		t.Fatal("expected non-nil store")
	}
	// Just verify it doesn't panic when adding entries.
	s.Add(newEntry("x"))
	if s.Len() != 1 {
		t.Fatal("expected 1 entry after add")
	}
}
