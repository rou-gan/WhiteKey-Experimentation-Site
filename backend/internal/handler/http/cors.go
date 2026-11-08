package http

import (
	"net/http"
	"os"
	"strings"
)

// CORS wraps h with Access-Control headers for browser callers (frontend).
// Allowed origins come from CORS_ALLOWED_ORIGINS (comma-separated).
// When unset, local Next.js defaults are used.
func CORS(h http.Handler) http.Handler {
	allowed := parseAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && originAllowed(origin, allowed) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func parseAllowedOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, strings.TrimRight(p, "/"))
		}
	}
	return out
}

func originAllowed(origin string, allowed []string) bool {
	origin = strings.TrimRight(origin, "/")
	for _, a := range allowed {
		if origin == a {
			return true
		}
	}
	return false
}
