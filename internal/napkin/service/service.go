package service

import (
	"context"
	"time"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

type Repository interface {
	Save(ctx context.Context, npk *napkin.Napkin, ttl time.Duration) error
	Get(ctx context.Context, code string) (*napkin.Napkin, error)
}

type Service struct {
	repo   Repository
	config Config
}

type Config struct {
	DefaultTTL       time.Duration
	CodeLength       int
	MaxContentLength int
}

func New(repo Repository, config Config) *Service {
	return &Service{repo: repo, config: config}
}
