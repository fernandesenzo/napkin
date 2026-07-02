package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fernandesenzo/napkin/internal/napkin"
	"github.com/fernandesenzo/napkin/internal/napkin/repository"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		repoNapkin  *napkin.Napkin
		repoErr     error
		expectedErr error
	}{
		{
			name:        "success",
			code:        "abcdef",
			repoNapkin:  &napkin.Napkin{Code: "abcdef", Text: "some text"},
			repoErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "invalid code length",
			code:        "abc",
			repoNapkin:  nil,
			repoErr:     nil,
			expectedErr: napkin.ErrInvalidCode,
		},
		{
			name:        "not found",
			code:        "abcdef",
			repoNapkin:  nil,
			repoErr:     repository.ErrNotFound,
			expectedErr: napkin.ErrNapkinDoesNotExist,
		},
		{
			name:        "repository error",
			code:        "abcdef",
			repoNapkin:  nil,
			repoErr:     errors.New("redis connection failed"),
			expectedErr: errors.New("service.Get: error getting napkin by code: redis connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRepository{
				getResult: tt.repoNapkin,
				getErr:    tt.repoErr,
			}
			svc := NewService(mock)

			got, err := svc.Get(context.Background(), tt.code)

			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				if !errors.Is(err, tt.expectedErr) && !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Errorf("expected error containing %q, got %q", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Code != tt.repoNapkin.Code || got.Text != tt.repoNapkin.Text {
				t.Errorf("expected napkin %+v, got %+v", tt.repoNapkin, got)
			}
		})
	}
}
