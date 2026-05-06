// Package fanout implements concurrent multi-target dispatch for retryq.
//
// A single enqueued item can carry more than one destination URL. The [Fan]
// function dispatches the same payload to every target simultaneously,
// honouring a per-target deadline, and collects the results into an ordered
// slice so callers can decide how to handle partial failures.
//
// Typical usage:
//
//	results := fanout.Fan(ctx, dispatcher, item.Targets, item.Headers, item.Body,
//		fanout.DefaultOptions())
//	if err := fanout.FirstError(results); err != nil {
//		// at least one target failed — schedule a retry
//	}
//
// The package is intentionally free of retry logic; retry decisions are
// delegated to the caller (usually internal/retry or internal/worker).
package fanout
