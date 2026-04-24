// Package metrics centralises all Prometheus instrumentation for retryq.
//
// Usage:
//
//	// Production — register against the default Prometheus registry.
//	 m := metrics.New(prometheus.DefaultRegisterer)
//
//	// Testing — use an isolated registry to avoid conflicts between tests.
//	 m := metrics.New(prometheus.NewRegistry())
//
// Counters
//
//	retryq_items_enqueued_total      – items accepted into the queue
//	retryq_items_retried_total       – individual retry attempts fired
//	retryq_items_dead_lettered_total – items that exhausted all attempts
//	retryq_items_succeeded_total     – items that eventually succeeded
//
// Gauges
//
//	retryq_queue_depth – live count of non-dead items currently queued
//
// Histograms
//
//	retryq_attempt_duration_seconds – latency of each outbound HTTP attempt
package metrics
