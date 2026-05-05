package shadowmode

import (
	"bytes"
	"io"
	"net/http"
)

// NewMiddleware returns an HTTP middleware that mirrors every request to the
// shadow target via d while passing the request unchanged to next.
func NewMiddleware(d *Dispatcher, next http.Handler) http.Handler {
	if d == nil || d.opts.ShadowURL == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bodyBytes []byte
		if r.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			// Restore body for the primary handler.
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		d.Mirror(r, bodyBytes)
		next.ServeHTTP(w, r)
	})
}
