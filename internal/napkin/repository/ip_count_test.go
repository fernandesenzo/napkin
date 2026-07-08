package repository

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestIPCounter(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	repo := NewRedisRepository(client)

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		ip            string
		action        string
		expectedCount int
		expectedErr   bool
	}{
		{
			name:          "GetEmpty",
			ip:            "127.0.0.1",
			action:        "get",
			expectedCount: 0,
			expectedErr:   false,
		},
		{
			name:          "IncrementNew",
			ip:            "127.0.0.1",
			action:        "increment",
			expectedCount: 1,
			expectedErr:   false,
		},
		{
			name: "IncrementExisting",
			setup: func(t *testing.T) {
				t.Helper()
				if err := s.Set("napking:ip:127.0.0.1", "5"); err != nil {
					t.Fatalf("miniredis Set: %v", err)
				}
			},
			ip:            "127.0.0.1",
			action:        "increment",
			expectedCount: 6,
			expectedErr:   false,
		},
		{
			name: "GetExisting",
			setup: func(t *testing.T) {
				t.Helper()
				if err := s.Set("napking:ip:127.0.0.1", "10"); err != nil {
					t.Fatalf("miniredis Set: %v", err)
				}
			},
			ip:            "127.0.0.1",
			action:        "get",
			expectedCount: 10,
			expectedErr:   false,
		},
		{
			name: "GetInvalidValue",
			setup: func(t *testing.T) {
				t.Helper()
				if err := s.Set("napking:ip:127.0.0.1", "not-an-int"); err != nil {
					t.Fatalf("miniredis Set: %v", err)
				}
			},
			ip:            "127.0.0.1",
			action:        "get",
			expectedCount: 0,
			expectedErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			if tt.setup != nil {
				tt.setup(t)
			}
			ctx := context.Background()
			if tt.action == "get" {
				count, err := repo.GetIPCounter(ctx, tt.ip)
				if (err != nil) != tt.expectedErr {
					t.Errorf("unexpected error state: %v", err)
				}
				if err == nil && count != tt.expectedCount {
					t.Errorf("expected count %d, got %d", tt.expectedCount, count)
				}
			} else if tt.action == "increment" {
				err := repo.IncrementIPCounter(ctx, tt.ip)
				if (err != nil) != tt.expectedErr {
					t.Errorf("unexpected error state: %v", err)
				}
				if err == nil {
					count, err := repo.GetIPCounter(ctx, tt.ip)
					if err != nil {
						t.Errorf("failed to get ip counter: %v", err)
					}
					if count != tt.expectedCount {
						t.Errorf("expected count %d, got %d", tt.expectedCount, count)
					}
					ttl := s.TTL("napking:ip:" + tt.ip)
					if ttl <= 0 {
						t.Errorf("expected ttl to be set, got %v", ttl)
					}
				}
			}
		})
	}
}
