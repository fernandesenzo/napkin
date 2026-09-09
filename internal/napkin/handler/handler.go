package handler

import (
	"context"
	"net/http"

	"github.com/fernandesenzo/napkin/internal/hub"
	"github.com/fernandesenzo/napkin/internal/napkin"
	"github.com/fernandesenzo/napkin/internal/origin"
	"github.com/gorilla/websocket"
)

type Service interface {
	Save(ctx context.Context, code string, content string) (*napkin.Napkin, error)
	Get(ctx context.Context, code string) (*napkin.Napkin, error)
}

type RoomManager interface {
	GetOrCreateRoom(code string) *hub.Hub
}

type Handler struct {
	svc        Service
	hubManager RoomManager
	config     Config
	upgrader   websocket.Upgrader
}

type Config struct {
	MaxContentLength int
	CodeLength       int
	AllowedOrigins   string
}

func New(svc Service, manager RoomManager, config Config) *Handler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  512,
		WriteBufferSize: 512,
		CheckOrigin: func(r *http.Request) bool {
			return origin.IsAllowed(r.Header.Get("Origin"), config.AllowedOrigins)
		},
	}

	return &Handler{
		svc:        svc,
		hubManager: manager,
		config:     config,
		upgrader:   upgrader,
	}
}
