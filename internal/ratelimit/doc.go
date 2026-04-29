// Package ratelimit implements a token-bucket rate limiter used by the
// retryq worker to throttle outbound dispatch requests.
//
// # Overview
//
// A [Limiter] is created with a sustained rate (tokens per second) and a
// burst capacity. Each call to [Limiter.Allow] attempts to consume one token;
// tokens refill continuously at the configured rate up to the burst cap.
//
// # Usage
//
//	l := ratelimit.New(ratelimit.Options{
//		Rate:  50,  // 50 req/s sustained
//		Burst: 100, // allow short spikes up to 100
//	})
//
//	if l.Allow() {
//		// dispatch the request
//	}
//
// A nil *Limiter is valid and always returns true from Allow, making it easy
// to disable rate limiting without changing call sites.
package ratelimit
