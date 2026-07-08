package manager

import (
	"context"
	"testing"

	"github.com/fernandesenzo/napkin/internal/client"
	"github.com/fernandesenzo/napkin/internal/napkin"
)

type mockSvc struct {
	saveFn func(ctx context.Context, code, content string) (*napkin.Napkin, error)
}

func (m *mockSvc) Save(ctx context.Context, code, content string) (*napkin.Napkin, error) {
	if m.saveFn != nil {
		return m.saveFn(ctx, code, content)
	}
	return nil, nil
}

func TestManager_GetOrCreateRoom(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{name: "creates new room", code: "aaaaaa"},
		{name: "creates room with different code", code: "bbbbbb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(&mockSvc{})
			h := m.GetOrCreateRoom(tt.code)
			if h == nil {
				t.Fatal("expected a non-nil hub")
			}
			if h.Code != tt.code {
				t.Errorf("expected hub code %q, got %q", tt.code, h.Code)
			}
		})
	}
}

func TestManager_GetOrCreateRoom_ReturnsSameHub(t *testing.T) {
	m := New(&mockSvc{})

	h1 := m.GetOrCreateRoom("aaaaaa")
	h2 := m.GetOrCreateRoom("aaaaaa")

	if h1 != h2 {
		t.Error("expected GetOrCreateRoom to return the same hub for the same code")
	}
}

func TestManager_GetOrCreateRoom_DifferentCodes(t *testing.T) {
	m := New(&mockSvc{})

	h1 := m.GetOrCreateRoom("aaaaaa")
	h2 := m.GetOrCreateRoom("bbbbbb")

	if h1 == h2 {
		t.Error("expected different hubs for different codes")
	}
}

func TestManager_RoomRemovedAfterEmpty(t *testing.T) {
	m := New(&mockSvc{})

	h := m.GetOrCreateRoom("aaaaaa")

	// Simulate a client joining then leaving via the Hub's exported channels.
	c := &client.Client{Send: make(chan string, 8)}
	h.RegisterChan() <- c

	// Trigger unregister (last client) — OnEmpty closes h and removes from map.
	h.UnregisterChan() <- c

	// Wait until OnEmpty has run (hub signals done via Close()).
	h.WaitDone()

	// A new call should create a fresh hub (different pointer).
	h2 := m.GetOrCreateRoom("aaaaaa")
	if h2 == h {
		t.Error("expected a new hub to be created after the room was removed")
	}
}
