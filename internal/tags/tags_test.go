package tags

import (
	"testing"
)

func TestSet_AddsTag(t *testing.T) {
	b := New()
	if err := b.Set("env", "prod"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tags := b.Build()
	if v, ok := tags.Get("env"); !ok || v != "prod" {
		t.Errorf("expected env=prod, got %q ok=%v", v, ok)
	}
}

func TestSet_EmptyKey_ReturnsError(t *testing.T) {
	b := New()
	if err := b.Set("", "value"); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestSet_ExceedsMaxTags_ReturnsError(t *testing.T) {
	b := New()
	for i := 0; i < MaxTags; i++ {
		if err := b.Set(fmt.Sprintf("k%d", i), "v"); err != nil {
			t.Fatalf("unexpected error at tag %d: %v", i, err)
		}
	}
	if err := b.Set("overflow", "x"); err != ErrTooManyTags {
		t.Errorf("expected ErrTooManyTags, got %v", err)
	}
}

func TestSet_OverwriteExistingKey_DoesNotCountAgain(t *testing.T) {
	b := New()
	for i := 0; i < MaxTags; i++ {
		_ = b.Set(fmt.Sprintf("k%d", i), "v")
	}
	// Overwriting an existing key must not trigger the limit.
	if err := b.Set("k0", "updated"); err != nil {
		t.Errorf("unexpected error overwriting existing key: %v", err)
	}
}

func TestMustSet_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	New().MustSet("", "bad")
}

func TestBuild_ReturnsCopy(t *testing.T) {
	b := New()
	_ = b.Set("x", "1")
	t1 := b.Build()
	_ = b.Set("y", "2")
	if _, ok := t1.Get("y"); ok {
		t.Error("Build snapshot should not reflect later mutations")
	}
}

func TestMatches_AllPresent_ReturnsTrue(t *testing.T) {
	tags := New().MustSet("env", "prod").MustSet("region", "us-east").Build()
	filter := New().MustSet("env", "prod").Build()
	if !tags.Matches(filter) {
		t.Error("expected Matches to return true")
	}
}

func TestMatches_MissingKey_ReturnsFalse(t *testing.T) {
	tags := New().MustSet("env", "prod").Build()
	filter := New().MustSet("region", "us-east").Build()
	if tags.Matches(filter) {
		t.Error("expected Matches to return false")
	}
}

func TestMatches_EmptyFilter_ReturnsTrue(t *testing.T) {
	tags := New().MustSet("env", "prod").Build()
	if !tags.Matches(Tags{}) {
		t.Error("empty filter should always match")
	}
}
