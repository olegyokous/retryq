package priority_test

import (
	"testing"

	"github.com/sashajdn/retryq/internal/priority"
)

func TestString_ReturnsExpectedLabel(t *testing.T) {
	cases := []struct {
		level priority.Level
		want  string
	}{
		{priority.High, "high"},
		{priority.Normal, "normal"},
		{priority.Low, "low"},
		{priority.Level(99), "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.level.String(); got != tc.want {
				t.Fatalf("String() = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestParse_KnownValues(t *testing.T) {
	cases := []struct {
		input string
		want  priority.Level
	}{
		{"high", priority.High},
		{"HIGH", priority.High},
		{"normal", priority.Normal},
		{"Normal", priority.Normal},
		{"", priority.Normal},
		{"low", priority.Low},
		{"LOW", priority.Low},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := priority.Parse(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("Parse(%q) = %v; want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParse_UnknownValue_ReturnsError(t *testing.T) {
	_, err := priority.Parse("critical")
	if err == nil {
		t.Fatal("expected error for unknown priority, got nil")
	}
}

func TestMustParse_ValidInput_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	_ = priority.MustParse("high")
}

func TestMustParse_InvalidInput_Panics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid input, got none")
		}
	}()
	priority.MustParse("ultra")
}

func TestLevel_Ordering(t *testing.T) {
	if priority.Low >= priority.Normal {
		t.Error("Low should be less than Normal")
	}
	if priority.Normal >= priority.High {
		t.Error("Normal should be less than High")
	}
}
