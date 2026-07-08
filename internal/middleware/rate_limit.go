package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/fernandesenzo/napkin/internal/ip"
	"github.com/redis/go-redis/v9"
)

func RateLimit(client *redis.Client, keyPrefix string, maxReq int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP, err := ip.ClientIP(r)
			if err != nil {
				slog.WarnContext(r.Context(), "middleware.RateLimit: failed to extract client ip, allowing request",
					"remote_addr", r.RemoteAddr,
					"err", err,
				)
				next.ServeHTTP(w, r)
				return
			}

			key := keyPrefix + clientIP

			count, err := client.Incr(r.Context(), key).Result()
			if err != nil {
				slog.ErrorContext(r.Context(), "middleware.RateLimit: redis INCR failed, allowing request",
					"ip", clientIP,
					"err", err,
				)
				next.ServeHTTP(w, r)
				return
			}
			if count == 1 {
				if err := client.Expire(r.Context(), key, window).Err(); err != nil {
					slog.WarnContext(r.Context(), "middleware.RateLimit: failed to set expiry on rate-limit key",
						"ip", clientIP,
						"err", err,
					)
				}
			}

			if count > int64(maxReq) {
				slog.WarnContext(r.Context(), "middleware.RateLimit: request blocked",
					"ip", clientIP,
					"count", count,
					"limit", maxReq,
					"key_prefix", keyPrefix,
				)
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
