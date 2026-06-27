package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/fernandesenzo/napkin/internal/napkin"
	"github.com/fernandesenzo/napkin/internal/napkin/repository"
)

func (s *Service) Get(ctx context.Context, code string) (*napkin.Napkin, error) {
	if err := napkin.ValidateCode(code); err != nil {
		return nil, err
	}
	npk, err := s.repo.Get(ctx, code)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("service.Get: error getting napkin by code: %w", err)
		}
		return nil, err
	}
	return npk, nil
}
