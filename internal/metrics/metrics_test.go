package metrics_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/yourorg/retryq/internal/metrics"
)

func newRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestNew_RegistersAllCollectors(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.New(reg)

	if m.Enqueued == nil {
		t.Error("Enqueued counter is nil")
	}
	if m.Retried == nil {
		t.Error("Retried counter is nil")
	}
	if m.DeadLettered == nil {
		t.Error("DeadLettered counter is nil")
	}
	if m.Succeeded == nil {
		t.Error("Succeeded counter is nil")
	}
	if m.QueueDepth == nil {
		t.Error("QueueDepth gauge is nil")
	}
	if m.AttemptDuration == nil {
		t.Error("AttemptDuration histogram is nil")
	}
}

func TestEnqueued_Increments(t *testing.T) {
	m := metrics.New(prometheus.NewRegistry())

	m.Enqueued.Add(3)

	if got := testutil.ToFloat64(m.Enqueued); got != 3 {
		t.Errorf("expected 3, got %v", got)
	}
}

func TestQueueDepth_SetAndDecrement(t *testing.T) {
	m := metrics.New(prometheus.NewRegistry())

	m.QueueDepth.Set(5)
	if got := testutil.ToFloat64(m.QueueDepth); got != 5 {
		t.Errorf("expected 5, got %v", got)
	}

	m.QueueDepth.Dec()
	if got := testutil.ToFloat64(m.QueueDepth); got != 4 {
		t.Errorf("expected 4 after Dec, got %v", got)
	}
}

func TestNew_PanicsOnDoubleRegister(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics.New(reg)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on double registration, got none")
		}
	}()

	metrics.New(reg)
}
