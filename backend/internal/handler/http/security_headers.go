package http

import "net/http"

// SecurityHeaders adds baseline response headers for API responses.
// Complements frontend CSP; APIs primarily need nosniff and framing denial.
func SecurityHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.ServeHTTP(w, r)
	})
}
