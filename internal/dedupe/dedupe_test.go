package dedupe_test

import (
	"testing"
	"time"

	"github.com/mcncl/retryq/internal/dedupe"
)

func TestIsDuplicate_NewIDReturnsFalse(t *testing.T) {
	s := dedupe.New(dedupe.DefaultOptions())
	if s.IsDuplicate("req-1") {
		t.Fatal("expected false for first-seen ID")
	}
}

func TestIsDuplicate_SameIDReturnsTrue(t *testing.T) {
	s := dedupe.New(dedupe.DefaultOptions())
	s.IsDuplicate("req-1")
	if !s.IsDuplicate("req-1") {
		t.Fatal("expected true for duplicate ID")
	}
}

func TestIsDuplicate_EmptyIDAlwaysFalse(t *testing.T) {
	s := dedupe.New(dedupe.DefaultOptions())
	if s.IsDuplicate("") {
		t.Fatal("empty ID should never be considered a duplicate")
	}
	if s.IsDuplicate("") {
		t.Fatal("empty ID should never be considered a duplicate on second call")
	}
}

func TestIsDuplicate_ExpiredIDAcceptedAgain(t *testing.T) {
	opts := dedupe.Options{
		TTL:     10 * time.Millisecond,
		MaxSize: 100,
	}
	s := dedupe.New(opts)
	s.IsDuplicate("req-expire")
	time.Sleep(20 * time.Millisecond)
	if s.IsDuplicate("req-expire") {
		t.Fatal("expected false after TTL expiry")
	}
}

func TestIsDuplicate_EvictsOldestWhenFull(t *testing.T) {
	opts := dedupe.Options{
		TTL:     1 * time.Hour,
		MaxSize: 3,
	}
	s := dedupe.New(opts)
	s.IsDuplicate("a")
	s.IsDuplicate("b")
	s.IsDuplicate("c")
	// Adding a 4th evicts "a".
	s.IsDuplicate("d")
	if s.Len() > 3 {
		t.Fatalf("expected at most 3 entries, got %d", s.Len())
	}
	// "a" should have been evicted and be accepted as new.
	if s.IsDuplicate("a") {
		t.Fatal("expected 'a' to be evicted and accepted as new")
	}
}

func TestLen_TracksEntries(t *testing.T) {
	s := dedupe.New(dedupe.DefaultOptions())
	if s.Len() != 0 {
		t.Fatalf("expected 0, got %d", s.Len())
	}
	s.IsDuplicate("x")
	s.IsDuplicate("y")
	if s.Len() != 2 {
		t.Fatalf("expected 2, got %d", s.Len())
	}
}

func TestNew_DefaultTTLAppliedWhenZero(t *testing.T) {
	// Providing a zero TTL should fall back to the default without panicking.
	s := dedupe.New(dedupe.Options{TTL: 0, MaxSize: 100})
	if s == nil {
		t.Fatal("expected non-nil store")
	}
	s.IsDuplicate("safe")
}
