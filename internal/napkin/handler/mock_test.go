package handler

import (
	"context"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

type mockService struct {
	saveFn               func(ctx context.Context, code string, content string) (*napkin.Napkin, error)
	incrementIPCounterFn func(ctx context.Context, ip string) error
	getFn                func(ctx context.Context, code string) (*napkin.Napkin, error)
}

func (m *mockService) Save(ctx context.Context, code string, content string) (*napkin.Napkin, error) {
	if m.saveFn != nil {
		return m.saveFn(ctx, code, content)
	}
	return nil, nil
}

func (m *mockService) IncrementIPCounter(ctx context.Context, ip string) error {
	if m.incrementIPCounterFn != nil {
		return m.incrementIPCounterFn(ctx, ip)
	}
	return nil
}

func (m *mockService) Get(ctx context.Context, code string) (*napkin.Napkin, error) {
	if m.getFn != nil {
		return m.getFn(ctx, code)
	}
	return nil, nil
}
