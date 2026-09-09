package service

import (
	"context"
	"fmt"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

func (s *Service) Save(ctx context.Context, code string, content string) (*napkin.Napkin, error) {
	if err := napkin.ValidateCode(code, s.config.CodeLength); err != nil {
		return nil, err
	}
	if err := napkin.ValidateContent(content, s.config.MaxContentLength); err != nil {
		return nil, err
	}
	npk := &napkin.Napkin{Code: code, Text: content}
	if err := s.repo.Save(ctx, npk, s.config.DefaultTTL); err != nil {
		return nil, fmt.Errorf("service.Save: failed to save napkin: %w", err)
	}
	return npk, nil
}
