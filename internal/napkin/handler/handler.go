package handler

import (
	"context"

	"github.com/fernandesenzo/napkin/internal/hub"
	"github.com/fernandesenzo/napkin/internal/napkin"
)

type Service interface {
	Save(ctx context.Context, code string, content string) (*napkin.Napkin, error)
	IncrementIPCounter(ctx context.Context, ip string) error
	Get(ctx context.Context, code string) (*napkin.Napkin, error)
}

type RoomManager interface {
	GetOrCreateRoom(code string) *hub.Hub
}

type Handler struct {
	svc        Service
	hubManager RoomManager
}

func New(svc Service, manager RoomManager) *Handler {
	return &Handler{svc: svc, hubManager: manager}
}
