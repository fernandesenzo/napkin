package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

// upgraderForTest accepts all origins.
var upgraderForTest = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// stubHub is a test double that satisfies the Hub interface.
type stubHub struct {
	code       string
	register   chan *Client
	unregister chan *Client
	broadcast  chan string
}

func newStubHub() *stubHub {
	return &stubHub{
		code:       "test01",
		register:   make(chan *Client, 4),
		unregister: make(chan *Client, 4),
		broadcast:  make(chan string, 16),
	}
}

func (s *stubHub) RegisterChan() chan<- *Client   { return s.register }
func (s *stubHub) UnregisterChan() chan<- *Client { return s.unregister }
func (s *stubHub) BroadcastChan() chan<- string   { return s.broadcast }
func (s *stubHub) GetCode() string                { return s.code }

// dialTestServer starts an httptest server that upgrades the connection and
// returns the server-side WebSocket conn and the client-side WebSocket conn.
func dialTestServer(t *testing.T, handler http.HandlerFunc) (*websocket.Conn, *websocket.Conn) {
	t.Helper()

	serverConnCh := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgraderForTest.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade error: %v", err)
			return
		}
		if handler != nil {
			handler(w, r)
		}
		serverConnCh <- conn
	}))
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { clientConn.Close() })

	serverConn := <-serverConnCh
	t.Cleanup(func() { serverConn.Close() })

	return serverConn, clientConn
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{name: "creates client with hub", code: "room01"},
		{name: "creates client with different code", code: "room02"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &stubHub{code: tt.code}
			c := NewClient(h, nil)
			if c == nil {
				t.Fatal("expected non-nil client")
			}
			if c.Send == nil {
				t.Fatal("expected Send channel to be initialized")
			}
			if cap(c.Send) != 256 {
				t.Errorf("expected Send capacity of 256, got %d", cap(c.Send))
			}
		})
	}
}

func TestStubHub_ImplementsHubInterface(t *testing.T) {
	// Compile-time check: stubHub satisfies Hub.
	var _ Hub = newStubHub()
}
