package manager

import (
	"context"
	"sync"

	"github.com/fernandesenzo/napkin/internal/hub"
	"github.com/fernandesenzo/napkin/internal/napkin"
)

type Service interface {
	Save(ctx context.Context, code string, content string) (*napkin.Napkin, error)
}
type Manager struct {
	mu    sync.Mutex
	rooms map[string]*hub.Hub
	svc   Service
}

func New(svc hub.Service) *Manager {
	return &Manager{
		rooms: make(map[string]*hub.Hub),
		svc:   svc,
	}
}

func (m *Manager) GetOrCreateRoom(code string) *hub.Hub {
	m.mu.Lock()
	defer m.mu.Unlock()

	if h, exists := m.rooms[code]; exists {
		return h
	}

	h := hub.New(code, m.svc)
	h.OnEmpty = func() {
		m.mu.Lock()
		delete(m.rooms, code)
		m.mu.Unlock()
		h.Close()
	}
	m.rooms[code] = h

	go h.Run()

	return h
}
