package service

import (
	"context"
	"time"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

type mockRepository struct {
	saveErr   error
	getResult *napkin.Napkin
	getErr    error
}

func (m *mockRepository) Save(_ context.Context, _ *napkin.Napkin, _ time.Duration) error {
	return m.saveErr
}

func (m *mockRepository) Get(_ context.Context, _ string) (*napkin.Napkin, error) {
	return m.getResult, m.getErr
}

