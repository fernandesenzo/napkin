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

func TestSave(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		body           string
		saveFn         func(ctx context.Context, code string, content string) (*napkin.Napkin, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "success",
			contentType: "application/json",
			body:        `{"code":"abcdef","content":"some text"}`,
			saveFn: func(ctx context.Context, code string, content string) (*napkin.Napkin, error) {
				return &napkin.Napkin{Code: code, Text: content}, nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"content":"some text"}`,
		},
		{
			name:           "unsupported content type",
			contentType:    "text/plain",
			body:           `{"code":"abcdef","content":"some text"}`,
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedBody:   "unsupported content type",
		},
		{
			name:           "invalid request json",
			contentType:    "application/json",
			body:           `{"code":`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request",
		},
		{
			name:           "unknown field",
			contentType:    "application/json",
			body:           `{"code":"abcdef","content":"some text","unknown":"value"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request",
		},
		{
			name:        "invalid code length",
			contentType: "application/json",
			body:        `{"code":"abc","content":"some text"}`,
			saveFn: func(ctx context.Context, code string, content string) (*napkin.Napkin, error) {
				return nil, napkin.ErrInvalidCode
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid code",
		},
		{
			name:        "content too long",
			contentType: "application/json",
			body:        `{"code":"abcdef","content":"` + strings.Repeat("a", 201) + `"}`,
			saveFn: func(ctx context.Context, code string, content string) (*napkin.Napkin, error) {
				return nil, napkin.ErrContentTooLong
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "content too long",
		},
		{
			name:        "internal server error",
			contentType: "application/json",
			body:        `{"code":"abcdef","content":"some text"}`,
			saveFn: func(ctx context.Context, code string, content string) (*napkin.Napkin, error) {
				return nil, errors.New("redis error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{
				saveFn: tt.saveFn,
			}
			h := New(svc, nil)

			req := httptest.NewRequest(http.MethodPost, "/napkin", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rr := httptest.NewRecorder()

			h.Save(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			body := strings.TrimSpace(rr.Body.String())
			if tt.expectedStatus == http.StatusCreated {
				var got saveNapkinResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				var want saveNapkinResponse
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
