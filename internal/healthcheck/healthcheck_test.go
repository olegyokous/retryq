package healthcheck_test

import (
	"testing"

	"github.com/sn3akiwhizper/retryq/internal/healthcheck"
)

func TestCheck_AllOK(t *testing.T) {
	a := healthcheck.New()
	a.Register(healthcheck.Checker{
		Name: "queue",
		Fn:   func() healthcheck.ComponentHealth { return healthcheck.ComponentHealth{Status: healthcheck.StatusOK} },
	})

	r := a.Check()
	if r.Status != healthcheck.StatusOK {
		t.Fatalf("expected ok, got %s", r.Status)
	}
	if _, ok := r.Components["queue"]; !ok {
		t.Fatal("expected queue component in report")
	}
}

func TestCheck_DegradedWinsOverOK(t *testing.T) {
	a := healthcheck.New()
	a.Register(healthcheck.Checker{
		Name: "alpha",
		Fn:   func() healthcheck.ComponentHealth { return healthcheck.ComponentHealth{Status: healthcheck.StatusOK} },
	})
	a.Register(healthcheck.Checker{
		Name: "beta",
		Fn:   func() healthcheck.ComponentHealth { return healthcheck.ComponentHealth{Status: healthcheck.StatusDegraded, Message: "high load"} },
	})

	r := a.Check()
	if r.Status != healthcheck.StatusDegraded {
		t.Fatalf("expected degraded, got %s", r.Status)
	}
}

func TestCheck_DownWinsOverDegraded(t *testing.T) {
	a := healthcheck.New()
	a.Register(healthcheck.Checker{
		Name: "x",
		Fn:   func() healthcheck.ComponentHealth { return healthcheck.ComponentHealth{Status: healthcheck.StatusDegraded} },
	})
	a.Register(healthcheck.Checker{
		Name: "y",
		Fn:   func() healthcheck.ComponentHealth { return healthcheck.ComponentHealth{Status: healthcheck.StatusDown, Message: "unreachable"} },
	})

	r := a.Check()
	if r.Status != healthcheck.StatusDown {
		t.Fatalf("expected down, got %s", r.Status)
	}
}

func TestCheck_NoCheckers_ReturnsOK(t *testing.T) {
	a := healthcheck.New()
	r := a.Check()
	if r.Status != healthcheck.StatusOK {
		t.Fatalf("expected ok with no checkers, got %s", r.Status)
	}
	if len(r.Components) != 0 {
		t.Fatalf("expected empty components, got %d", len(r.Components))
	}
}

func TestCheck_TimestampSet(t *testing.T) {
	a := healthcheck.New()
	r := a.Check()
	if r.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}
