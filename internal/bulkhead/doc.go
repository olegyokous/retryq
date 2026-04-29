// Package bulkhead provides a concurrency-limiting bulkhead for HTTP services.
//
// A bulkhead caps the number of simultaneous in-flight requests processed by
// a handler, preventing thread/goroutine exhaustion when a downstream service
// is slow or unavailable. When the cap is reached the middleware responds
// immediately with HTTP 503 so callers can back off or retry via retryq.
//
// # Usage
//
//	b := bulkhead.New(bulkhead.DefaultOptions())
//	http.Handle("/enqueue", bulkhead.NewMiddleware(b, myHandler))
//
// # Disabling
//
// Set MaxConcurrent to 0 (or use a nil *Bulkhead in NewMiddleware) to
// disable the limit entirely — useful in tests or local development.
package bulkhead
