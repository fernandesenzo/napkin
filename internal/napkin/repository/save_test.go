package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/fernandesenzo/napkin/internal/napkin"
	"github.com/redis/go-redis/v9"
)

func TestSave(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	repo := NewRedisRepository(client)

	tests := []struct {
		name        string
		napkin      *napkin.Napkin
		ttl         time.Duration
		expectedErr bool
	}{
		{
			name: "Success",
			napkin: &napkin.Napkin{
				Code: "abc",
				Text: "some text",
			},
			ttl:         time.Minute,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			err := repo.Save(context.Background(), tt.napkin, tt.ttl)
			if (err != nil) != tt.expectedErr {
				t.Errorf("unexpected error state: %v", err)
			}
			if err == nil {
				val, err := s.Get("napkin:napkin:" + tt.napkin.Code)
				if err != nil {
					t.Errorf("expected key to be set in redis: %v", err)
				}
				if val != tt.napkin.Text {
					t.Errorf("expected stored text %q, got %q", tt.napkin.Text, val)
				}
				ttl := s.TTL("napkin:napkin:" + tt.napkin.Code)
				if ttl <= 0 {
					t.Errorf("expected ttl to be set, got %v", ttl)
				}
			}
		})
	}
}
