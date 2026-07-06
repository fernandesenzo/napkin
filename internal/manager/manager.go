package manager

import (
	"sync"

	"github.com/fernandesenzo/napkin/internal/hub"
)

type Manager struct {
	mu    sync.Mutex
	rooms map[string]*hub.Hub
	svc   hub.Service
}

func NewManager(svc hub.Service) *Manager {
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

	h := hub.NewHub(code, m.svc)
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
