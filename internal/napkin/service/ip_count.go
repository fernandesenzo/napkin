package service

import (
	"context"
	"fmt"
)

func (s *Service) IncrementIPCounter(ctx context.Context, ip string) error {
	if err := s.repo.IncrementIPCounter(ctx, ip); err != nil {
		return fmt.Errorf("service.IncrementIPCounter: failed to increment ip counter: %w", err)
	}
	return nil
}
