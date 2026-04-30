// Package sampling provides probabilistic sampling primitives for retryq.
//
// # Overview
//
// A [Sampler] is constructed with a rate in [0,1] and a deterministic RNG
// seed. Calling [Sampler.Sample] returns true with probability equal to the
// configured rate, making it straightforward to shed load or reduce audit
// verbosity during high-throughput periods.
//
// # Middleware
//
// [NewMiddleware] wraps any http.Handler and drops requests that are not
// sampled, responding with HTTP 429 and a Retry-After hint so well-behaved
// clients can back off gracefully.
//
// # Thread safety
//
// Sampler is safe for concurrent use; the internal RNG is protected by a
// mutex.
package sampling
