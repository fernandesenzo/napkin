package hub

import (
	"context"

	"github.com/fernandesenzo/napkin/internal/client"
	"github.com/fernandesenzo/napkin/internal/napkin"
)

type Service interface {
	Save(ctx context.Context, code string, content string) (*napkin.Napkin, error)
}

type Hub struct {
	Code       string
	svc        Service
	clients    map[*client.Client]bool
	broadcast  chan string
	register   chan *client.Client
	unregister chan *client.Client
	done       chan struct{}
	OnEmpty    func()
}

func NewHub(code string, svc Service) *Hub {
	return &Hub{
		Code:       code,
		svc:        svc,
		clients:    make(map[*client.Client]bool),
		broadcast:  make(chan string),
		register:   make(chan *client.Client),
		unregister: make(chan *client.Client),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Join(c *client.Client) bool {
	select {
	case h.register <- c:
		return true
	case <-h.done:
		return false
	}
}

func (h *Hub) RegisterChan() chan<- *client.Client {
	return h.register
}

func (h *Hub) UnregisterChan() chan<- *client.Client {
	return h.unregister
}

func (h *Hub) BroadcastChan() chan<- string {
	return h.broadcast
}

func (h *Hub) GetCode() string {
	return h.Code
}

// Close signals that the hub is done by closing the done channel.
func (h *Hub) Close() {
	close(h.done)
}

// WaitDone blocks until the hub is closed.
func (h *Hub) WaitDone() {
	<-h.done
}
