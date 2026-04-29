package deadletter

import (
	"testing"
	"time"
)

func makeExpiredEntry(id string, age time.Duration) Entry {
	return Entry{
		ID:       id,
		FailedAt: time.Now().Add(-age),
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	s := New(10)
	s.Add(makeExpiredEntry("old1", 5*time.Hour))
	s.Add(makeExpiredEntry("old2", 4*time.Hour))
	s.Add(makeExpiredEntry("fresh", 1*time.Minute))

	removed := Purge(s, 2*time.Hour)

	if removed != 2 {
		t.Fatalf("expected 2 removed, got %d", removed)
	}
	if s.Len() != 1 {
		t.Fatalf("expected 1 remaining, got %d", s.Len())
	}
}

func TestPurge_ZeroMaxAge_IsNoop(t *testing.T) {
	s := New(10)
	s.Add(makeExpiredEntry("old", 100*time.Hour))

	removed := Purge(s, 0)

	if removed != 0 {
		t.Fatalf("expected 0 removed, got %d", removed)
	}
	if s.Len() != 1 {
		t.Fatal("entry should still be present")
	}
}

func TestPurge_AllFresh_RemovesNone(t *testing.T) {
	s := New(10)
	s.Add(makeExpiredEntry("a", 30*time.Second))
	s.Add(makeExpiredEntry("b", 45*time.Second))

	removed := Purge(s, 2*time.Hour)

	if removed != 0 {
		t.Fatalf("expected 0 removed, got %d", removed)
	}
	if s.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", s.Len())
	}
}

func TestDefaultExpiryOptions_MaxAge(t *testing.T) {
	opts := DefaultExpiryOptions()
	if opts.MaxAge != 72*time.Hour {
		t.Fatalf("expected 72h, got %v", opts.MaxAge)
	}
}
