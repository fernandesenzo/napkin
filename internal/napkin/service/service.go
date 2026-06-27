package service

import (
	"context"
	"time"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

type Repository interface {
	Save(ctx context.Context, npk *napkin.Napkin, ttl time.Duration) error
	IncrementIPCounter(ctx context.Context, ip string) error
	Get(ctx context.Context, code string) (*napkin.Napkin, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}
