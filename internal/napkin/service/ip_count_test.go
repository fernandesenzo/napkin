package service

import (
	"context"
	"errors"
	"testing"
)

func TestIncrementIPCounter(t *testing.T) {
	tests := []struct {
		name        string
		ip          string
		repoErr     error
		expectedErr bool
	}{
		{
			name:        "success",
			ip:          "192.168.0.1",
			repoErr:     nil,
			expectedErr: false,
		},
		{
			name:        "repository error",
			ip:          "192.168.0.1",
			repoErr:     errors.New("redis connection failed"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRepository{
				incrementIPErr: tt.repoErr,
			}
			svc := New(mock)

			err := svc.IncrementIPCounter(context.Background(), tt.ip)

			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error=%v, got %v", tt.expectedErr, err)
			}
		})
	}
}
