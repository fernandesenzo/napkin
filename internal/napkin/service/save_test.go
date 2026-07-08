package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

func TestSave(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		content     string
		repoErr     error
		expectedErr error
	}{
		{
			name:        "success",
			code:        "abcdef",
			content:     "some text",
			repoErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "invalid code length",
			code:        "abc",
			content:     "some text",
			repoErr:     nil,
			expectedErr: napkin.ErrInvalidCode,
		},
		{
			name:        "repository error",
			code:        "abcdef",
			content:     "some text",
			repoErr:     errors.New("redis connection failed"),
			expectedErr: errors.New("service.Save: failed to save napkin: redis connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRepository{
				saveErr: tt.repoErr,
			}
			svc := New(mock)

			got, err := svc.Save(context.Background(), tt.code, tt.content)

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

			if got.Code != tt.code || got.Text != tt.content {
				t.Errorf("expected napkin {Code: %q, Text: %q}, got %+v", tt.code, tt.content, got)
			}
		})
	}
}
