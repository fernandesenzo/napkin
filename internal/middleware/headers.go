package middleware

import (
	"net/http"

	"github.com/fernandesenzo/napkin/internal/origin"
)

func ApplyHeaders(allowedOrigins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			if allowedOrigins == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if allowedOrigins != "" {
				reqOrigin := r.Header.Get("Origin")
				if origin.IsAllowed(reqOrigin, allowedOrigins) {
					w.Header().Set("Access-Control-Allow-Origin", reqOrigin)
					w.Header().Add("Vary", "Origin")
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
