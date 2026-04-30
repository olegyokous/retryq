package sampling

import (
	"net/http"
	"strconv"
)

// NewMiddleware returns an HTTP middleware that gates requests through the
// sampler. Requests that are not sampled receive 429 with a Retry-After
// header derived from the configured rate so callers can back off.
//
// A nil sampler passes every request through unchanged.
func NewMiddleware(s *Sampler, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s == nil || s.Sample() {
			next.ServeHTTP(w, r)
			return
		}
		// Suggest a back-off inversely proportional to the sample rate.
		retryAfter := retryAfterSeconds(s.Rate())
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		w.Header().Set("X-Sample-Rate", strconv.FormatFloat(s.Rate(), 'f', 4, 64))
		http.Error(w, "request not sampled", http.StatusTooManyRequests)
	})
}

// retryAfterSeconds returns a suggested retry delay in seconds.
func retryAfterSeconds(rate float64) int {
	if rate <= 0 {
		return 60
	}
	v := int(1.0 / rate)
	if v < 1 {
		return 1
	}
	return v
}
