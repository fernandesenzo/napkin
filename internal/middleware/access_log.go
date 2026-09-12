package middleware

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/fernandesenzo/napkin/internal/ip"
)

type responseWriterObserver struct {
	http.ResponseWriter
	status int
}

func (w *responseWriterObserver) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterObserver) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *responseWriterObserver) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	return hj.Hijack()
}

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		observer := &responseWriterObserver{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(observer, r)

		clientIP, _ := ip.ClientIP(r)

		slog.InfoContext(r.Context(), "request finished",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("ip", clientIP),
			slog.Int("status", observer.status),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
		)
	})
}
