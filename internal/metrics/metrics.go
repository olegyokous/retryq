// Package metrics provides Prometheus instrumentation for retryq.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "retryq"

// Metrics holds all Prometheus collectors for the retry queue.
type Metrics struct {
	Enqueued    prometheus.Counter
	Retried     prometheus.Counter
	DeadLettered prometheus.Counter
	Succeeded   prometheus.Counter
	QueueDepth  prometheus.Gauge
	AttemptDuration prometheus.Histogram
}

// New registers and returns a new Metrics instance using the provided registerer.
// Pass prometheus.DefaultRegisterer for production use.
func New(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)

	return &Metrics{
		Enqueued: factory.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "items_enqueued_total",
			Help:      "Total number of items added to the retry queue.",
		}),
		Retried: factory.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "items_retried_total",
			Help:      "Total number of retry attempts made.",
		}),
		DeadLettered: factory.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "items_dead_lettered_total",
			Help:      "Total number of items moved to the dead-letter queue.",
		}),
		Succeeded: factory.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "items_succeeded_total",
			Help:      "Total number of items that succeeded on retry.",
		}),
		QueueDepth: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "queue_depth",
			Help:      "Current number of active (non-dead) items in the queue.",
		}),
		AttemptDuration: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "attempt_duration_seconds",
			Help:      "Duration of individual retry attempt HTTP calls.",
			Buckets:   prometheus.DefBuckets,
		}),
	}
}
