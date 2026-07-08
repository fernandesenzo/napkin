package handler

import (
	"context"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

type mockService struct {
	saveFn func(ctx context.Context, code string, content string) (*napkin.Napkin, error)
	getFn  func(ctx context.Context, code string) (*napkin.Napkin, error)
}

func (m *mockService) Save(ctx context.Context, code string, content string) (*napkin.Napkin, error) {
	if m.saveFn != nil {
		return m.saveFn(ctx, code, content)
	}
	return nil, nil
}

func (m *mockService) Get(ctx context.Context, code string) (*napkin.Napkin, error) {
	if m.getFn != nil {
		return m.getFn(ctx, code)
	}
	return nil, nil
}
