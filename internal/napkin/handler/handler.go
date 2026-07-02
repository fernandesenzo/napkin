package handler

import (
	"context"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

type Service interface {
	Save(ctx context.Context, code string, content string) (*napkin.Napkin, error)
	IncrementIPCounter(ctx context.Context, ip string) error
	Get(ctx context.Context, code string) (*napkin.Napkin, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}
