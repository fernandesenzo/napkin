package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		getFn          func(ctx context.Context, code string) (*napkin.Napkin, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			code: "abcdef",
			getFn: func(ctx context.Context, code string) (*napkin.Napkin, error) {
				return &napkin.Napkin{Code: "abcdef", Text: "some content"}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"code":"abcdef","content":"some content"}`,
		},
		{
			name: "invalid code length",
			code: "abc",
			getFn: func(ctx context.Context, code string) (*napkin.Napkin, error) {
				return nil, napkin.ErrInvalidCode
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid code",
		},
		{
			name: "not found",
			code: "abcdef",
			getFn: func(ctx context.Context, code string) (*napkin.Napkin, error) {
				return nil, napkin.ErrNapkinDoesNotExist
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "napkin not found",
		},
		{
			name: "internal server error",
			code: "abcdef",
			getFn: func(ctx context.Context, code string) (*napkin.Napkin, error) {
				return nil, errors.New("database connection failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{
				getFn: tt.getFn,
			}
			h := New(svc, nil)

			req := httptest.NewRequest(http.MethodGet, "/napkin/"+tt.code, nil)
			req.SetPathValue("code", tt.code)
			rr := httptest.NewRecorder()

			h.Get(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			body := strings.TrimSpace(rr.Body.String())
			if tt.expectedStatus == http.StatusOK {
				var got getNapkinResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				var want getNapkinResponse
				if err := json.Unmarshal([]byte(tt.expectedBody), &want); err != nil {
					t.Fatalf("failed to unmarshal expected body: %v", err)
				}
				if got != want {
					t.Errorf("expected body %v, got %v", want, got)
				}
			} else {
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}
		})
	}
}
