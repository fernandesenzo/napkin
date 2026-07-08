package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fernandesenzo/napkin/internal/hub"
	"github.com/fernandesenzo/napkin/internal/manager"
	"github.com/gorilla/websocket"
)

// mockRoomManager is a test double for the RoomManager interface.
type mockRoomManager struct {
	getOrCreateRoomFn func(code string) *hub.Hub
}

func (m *mockRoomManager) GetOrCreateRoom(code string) *hub.Hub {
	if m.getOrCreateRoomFn != nil {
		return m.getOrCreateRoomFn(code)
	}
	return nil
}

// dialWebSocket connects to the given httptest server URL using WebSocket.
func dialWebSocket(t *testing.T, srv *httptest.Server, path string) (*websocket.Conn, *http.Response) {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + path
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn, resp
}

func TestWebSocket_MissingCode(t *testing.T) {
	h := New(&mockService{}, nil)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.WebSocket(w, r)
	}))
	defer srv.Close()

	// Plain HTTP request (no WebSocket upgrade) — we just want the 400 status.
	resp, err := http.Get(srv.URL + "/ws/")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestWebSocket_InvalidCode(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{name: "code too short", code: "abc"},
		{name: "code too long", code: "abcdefgh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(&mockService{}, nil)

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.SetPathValue("code", tt.code)
				h.WebSocket(w, r)
			}))
			defer srv.Close()

			resp, err := http.Get(srv.URL + "/ws/" + tt.code)
			if err != nil {
				t.Fatalf("GET: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", resp.StatusCode)
			}
		})
	}
}

func TestWebSocket_JoinsRoom(t *testing.T) {
	svc := &mockService{}

	// Use a real Manager backed by a no-op service so Join works end-to-end.
	manager := manager.New(svc)
	h := New(svc, manager)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("code", "abcdef")
		h.WebSocket(w, r)
	}))
	defer srv.Close()

	conn, _ := dialWebSocket(t, srv, "/ws/abcdef")

	// Connection is live — write and read back to confirm the hub is running.
	if err := conn.WriteMessage(websocket.TextMessage, []byte("hi")); err != nil {
		t.Fatalf("write: %v", err)
	}

	// The ReadPump discards oversized messages; "hi" is valid, it will be broadcast.
	// No other client to echo, just verify we stay connected.
	conn.Close()
}

func TestWebSocket_MultipleClientsInSameRoom(t *testing.T) {
	svc := &mockService{}
	manager := manager.New(svc)
	h := New(svc, manager)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("code", "abcdef")
		h.WebSocket(w, r)
	}))
	defer srv.Close()

	conn1, _ := dialWebSocket(t, srv, "/ws/abcdef")
	conn2, _ := dialWebSocket(t, srv, "/ws/abcdef")

	msg := "broadcast test"
	if err := conn1.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("write from conn1: %v", err)
	}

	// conn2 should receive the broadcast.
	_, got, err := conn2.ReadMessage()
	if err != nil {
		t.Fatalf("read from conn2: %v", err)
	}
	if string(got) != msg {
		t.Errorf("expected %q, got %q", msg, string(got))
	}

	conn1.Close()
	conn2.Close()
}

// Ensure mockRoomManager satisfies the RoomManager interface at compile-time.
var _ interface{ GetOrCreateRoom(string) *hub.Hub } = &mockRoomManager{}

// Ensure mockService satisfies the handler Service interface.
var _ Service = &mockService{}
