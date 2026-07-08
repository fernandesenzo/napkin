package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/fernandesenzo/napkin/internal/middleware"
	"github.com/redis/go-redis/v9"
)

func newTestClient(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	return s, client
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimit(t *testing.T) {
	const prefix = "napkin:rl:test:"
	const limit = 3
	const window = time.Minute

	t.Run("allows requests under the limit", func(t *testing.T) {
		_, client := newTestClient(t)
		m := middleware.RateLimit(client, prefix, limit, window)(okHandler())

		for i := range limit {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "127.0.0.1:1234"
			rec := httptest.NewRecorder()
			m.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("request %d: got status %d, want %d", i+1, rec.Code, http.StatusOK)
			}
		}
	})

	t.Run("blocks request that exceeds the limit", func(t *testing.T) {
		_, client := newTestClient(t)
		m := middleware.RateLimit(client, prefix, limit, window)(okHandler())

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"

		for range limit {
			rec := httptest.NewRecorder()
			m.ServeHTTP(rec, req)
		}

		rec := httptest.NewRecorder()
		m.ServeHTTP(rec, req)
		if rec.Code != http.StatusTooManyRequests {
			t.Errorf("got status %d, want %d", rec.Code, http.StatusTooManyRequests)
		}
	})

	t.Run("counters are isolated by key prefix", func(t *testing.T) {
		_, client := newTestClient(t)
		mA := middleware.RateLimit(client, "napkin:rl:a:", 1, window)(okHandler())
		mB := middleware.RateLimit(client, "napkin:rl:b:", 1, window)(okHandler())

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"

		// exhaust prefix A
		for range 2 {
			rec := httptest.NewRecorder()
			mA.ServeHTTP(rec, req)
		}

		// prefix B should still allow requests
		rec := httptest.NewRecorder()
		mB.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("isolated prefix B got status %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("counters are isolated by IP", func(t *testing.T) {
		_, client := newTestClient(t)
		m := middleware.RateLimit(client, prefix, 1, window)(okHandler())

		reqA := httptest.NewRequest(http.MethodGet, "/", nil)
		reqA.RemoteAddr = "10.0.0.1:1234"

		reqB := httptest.NewRequest(http.MethodGet, "/", nil)
		reqB.RemoteAddr = "10.0.0.2:1234"

		// exhaust IP A
		for range 2 {
			rec := httptest.NewRecorder()
			m.ServeHTTP(rec, reqA)
		}

		// IP B should not be affected
		rec := httptest.NewRecorder()
		m.ServeHTTP(rec, reqB)
		if rec.Code != http.StatusOK {
			t.Errorf("isolated IP B got status %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("counter resets after window expires", func(t *testing.T) {
		s, client := newTestClient(t)
		m := middleware.RateLimit(client, prefix, 1, window)(okHandler())

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"

		// exhaust the limit
		for range 2 {
			rec := httptest.NewRecorder()
			m.ServeHTTP(rec, req)
		}

		// fast-forward past the window in miniredis
		s.FastForward(window + time.Second)

		rec := httptest.NewRecorder()
		m.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("after window reset got status %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("fails open when RemoteAddr is invalid", func(t *testing.T) {
		_, client := newTestClient(t)
		m := middleware.RateLimit(client, prefix, limit, window)(okHandler())

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "invalid-addr"
		rec := httptest.NewRecorder()

		m.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
		}
	})
}
