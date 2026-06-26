package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/fernandesenzo/napkin/internal/napkin"
	"github.com/redis/go-redis/v9"
)

func TestGet(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	repo := NewRedisRepository(client)

	tests := []struct {
		name        string
		setup       func()
		code        string
		expected    *napkin.Napkin
		expectedErr error
	}{
		{
			name: "Success",
			setup: func() {
				s.Set("napkin:napkin:abc", "some content")
			},
			code: "abc",
			expected: &napkin.Napkin{
				Code: "abc",
				Text: "some content",
			},
			expectedErr: nil,
		},
		{
			name:        "NotFound",
			code:        "xyz",
			expected:    nil,
			expectedErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			if tt.setup != nil {
				tt.setup()
			}
			res, err := repo.Get(context.Background(), tt.code)
			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if res == nil || res.Code != tt.expected.Code || res.Text != tt.expected.Text {
					t.Errorf("expected %+v, got %+v", tt.expected, res)
				}
			}
		})
	}
}
