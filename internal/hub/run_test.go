package hub

import (
	"context"
	"sync"
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

// newTestHub returns a Hub backed by a no-op service.
func newTestHub() *Hub {
	return NewHub("test01", &mockSvc{})
}

func TestHub_RegisterClient(t *testing.T) {
	h := newTestHub()
	go h.Run()

	c := &client.Client{Send: make(chan string, 8)}

	if !h.Join(c) {
		t.Fatal("Join should return true when hub is alive")
	}
}

func TestHub_JoinReturnsFalseWhenDone(t *testing.T) {
	h := newTestHub()
	close(h.done)

	c := &client.Client{Send: make(chan string, 8)}
	if h.Join(c) {
		t.Fatal("Join should return false when hub is done")
	}
}

func TestHub_UnregisterLastClientCallsOnEmpty(t *testing.T) {
	h := newTestHub()

	var mu sync.Mutex
	called := false
	h.OnEmpty = func() {
		mu.Lock()
		defer mu.Unlock()
		called = true
		h.Close()
	}

	go h.Run()

	c := &client.Client{Send: make(chan string, 8)}
	h.register <- c

	h.unregister <- c

	// Wait for the hub goroutine to finish
	h.WaitDone()

	mu.Lock()
	defer mu.Unlock()
	if !called {
		t.Error("expected OnEmpty to be called when the last client leaves")
	}
}

func TestHub_BroadcastSendsToAllClients(t *testing.T) {
	h := newTestHub()
	go h.Run()

	c1 := &client.Client{Send: make(chan string, 8)}
	c2 := &client.Client{Send: make(chan string, 8)}

	h.register <- c1
	h.register <- c2

	h.broadcast <- "hello"

	got1 := <-c1.Send
	got2 := <-c2.Send

	if got1 != "hello" {
		t.Errorf("c1: expected 'hello', got %q", got1)
	}
	if got2 != "hello" {
		t.Errorf("c2: expected 'hello', got %q", got2)
	}
}

func TestHub_UnregisterUnknownClientIsNoop(t *testing.T) {
	h := newTestHub()
	go h.Run()

	c1 := &client.Client{Send: make(chan string, 8)}
	c2 := &client.Client{Send: make(chan string, 8)}

	h.register <- c1

	// Unregistering c2 (not in the hub) should be a no-op.
	h.unregister <- c2

	// c1 should still receive broadcasts.
	h.broadcast <- "still alive"
	got := <-c1.Send
	if got != "still alive" {
		t.Errorf("expected 'still alive', got %q", got)
	}
}
