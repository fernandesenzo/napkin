package e2e

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/fernandesenzo/napkin/internal/manager"
	"github.com/fernandesenzo/napkin/internal/middleware"
	"github.com/fernandesenzo/napkin/internal/napkin/handler"
	"github.com/fernandesenzo/napkin/internal/napkin/repository"
	"github.com/fernandesenzo/napkin/internal/napkin/service"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

func newTestServer(t *testing.T) (string, func()) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	repo := repository.NewRedisRepository(client)
	svc := service.New(repo)
	mgr := manager.New(svc)
	h := handler.New(svc, mgr)

	rlGet := middleware.RateLimit(client, "napkin:rl:get:", 100, time.Minute)
	rlPost := middleware.RateLimit(client, "napkin:rl:post:", 5, time.Minute)

	mux := http.NewServeMux()
	mux.Handle("POST /api/save", middleware.BodyLimit(4096)(rlPost(http.HandlerFunc(h.Save))))
	mux.Handle("GET /{code}", rlGet(http.HandlerFunc(h.Get)))
	mux.Handle("GET /{code}/ws", rlGet(http.HandlerFunc(h.WebSocket)))

	var handlerStack http.Handler = mux
	handlerStack = middleware.ApplyHeaders("*")(handlerStack)
	handlerStack = middleware.InjectReqID(handlerStack)
	handlerStack = middleware.Recover(handlerStack)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	srv := &http.Server{Handler: handlerStack}
	go func() { _ = srv.Serve(listener) }()

	addr := "http://" + listener.Addr().String()
	cleanup := func() {
		srv.Close()
		mr.Close()
	}
	return addr, cleanup
}

func TestSaveNapkin(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	body := `{"code":"abc123","content":"hello world"}`
	resp, err := http.Post(addr+"/api/save", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var result map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&result)
	if result["content"] != "hello world" {
		t.Errorf("expected content 'hello world', got %q", result["content"])
	}
}

func TestGetNapkin(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	saveBody := `{"code":"xyz789","content":"test content"}`
	resp, _ := http.Post(addr+"/api/save", "application/json", strings.NewReader(saveBody))
	resp.Body.Close()

	resp2, err := http.Get(addr + "/xyz789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp2.StatusCode)
	}

	var result map[string]string
	_ = json.NewDecoder(resp2.Body).Decode(&result)
	if result["code"] != "xyz789" {
		t.Errorf("expected code 'xyz789', got %q", result["code"])
	}
	if result["content"] != "test content" {
		t.Errorf("expected content 'test content', got %q", result["content"])
	}
}

func TestGetNonexistentNapkin(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	resp, err := http.Get(addr + "/ZZZZZZ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestSaveInvalidCode(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	body := `{"code":"ab","content":"hello"}`
	resp, err := http.Post(addr+"/api/save", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSaveContentTooLong(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	longContent := strings.Repeat("a", 201)
	body := `{"code":"abc123","content":"` + longContent + `"}`
	resp, err := http.Post(addr+"/api/save", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}
}

func TestSaveWrongContentType(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	resp, err := http.Post(addr+"/api/save", "text/plain", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", resp.StatusCode)
	}
}

func TestOverwriteNapkin(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	firstBody := `{"code":"ovr123","content":"original"}`
	resp, _ := http.Post(addr+"/api/save", "application/json", strings.NewReader(firstBody))
	resp.Body.Close()

	secondBody := `{"code":"ovr123","content":"updated"}`
	resp2, _ := http.Post(addr+"/api/save", "application/json", strings.NewReader(secondBody))
	resp2.Body.Close()

	resp3, err := http.Get(addr + "/ovr123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp3.Body.Close()

	var result map[string]string
	_ = json.NewDecoder(resp3.Body).Decode(&result)
	if result["content"] != "updated" {
		t.Errorf("expected content 'updated', got %q", result["content"])
	}
}

func TestWebSocketConnectAndEcho(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	saveBody := `{"code":"wsroom","content":"initial"}`
	resp, _ := http.Post(addr+"/api/save", "application/json", strings.NewReader(saveBody))
	resp.Body.Close()

	wsURL := "ws" + strings.TrimPrefix(addr, "http") + "/wsroom/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer conn.Close()

	msg := "hello from ws"
	if writeErr := conn.WriteMessage(websocket.TextMessage, []byte(msg)); writeErr != nil {
		t.Fatalf("failed to write message: %v", writeErr)
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if string(data) != msg {
		t.Errorf("expected %q, got %q", msg, string(data))
	}
}

func TestWebSocketBroadcast(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	saveBody := `{"code":"broadc","content":"initial"}`
	resp, _ := http.Post(addr+"/api/save", "application/json", strings.NewReader(saveBody))
	resp.Body.Close()

	wsURL := "ws" + strings.TrimPrefix(addr, "http") + "/broadc/ws"

	connA, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client A: failed to dial: %v", err)
	}
	defer connA.Close()

	connB, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client B: failed to dial: %v", err)
	}
	defer connB.Close()

	msg := "broadcast test"
	if writeErr := connA.WriteMessage(websocket.TextMessage, []byte(msg)); writeErr != nil {
		t.Fatalf("client A: failed to write: %v", writeErr)
	}

	_ = connB.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, dataB, err := connB.ReadMessage()
	if err != nil {
		t.Fatalf("client B: failed to read: %v", err)
	}
	if string(dataB) != msg {
		t.Errorf("client B expected %q, got %q", msg, string(dataB))
	}

	_ = connA.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, dataA, err := connA.ReadMessage()
	if err != nil {
		t.Fatalf("client A: failed to read: %v", err)
	}
	if string(dataA) != msg {
		t.Errorf("client A expected %q, got %q", msg, string(dataA))
	}
}

func TestWebSocketInvalidCode(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	wsURL := "ws" + strings.TrimPrefix(addr, "http") + "/ab/ws"
	_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected websocket upgrade to fail for invalid code")
	}
}

func TestWebSocketMessageTooLong(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	saveBody := `{"code":"toolng","content":"initial"}`
	resp, _ := http.Post(addr+"/api/save", "application/json", strings.NewReader(saveBody))
	resp.Body.Close()

	wsURL := "ws" + strings.TrimPrefix(addr, "http") + "/toolng/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	longMsg := strings.Repeat("x", 201)
	if writeErr := conn.WriteMessage(websocket.TextMessage, []byte(longMsg)); writeErr != nil {
		t.Fatalf("failed to write: %v", writeErr)
	}

	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Error("expected no broadcast for oversized message")
	}
}

func TestPostRateLimit(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		code := "rlt" + strings.Repeat("0", 2) + string(rune('0'+i))
		body := `{"code":"` + code + `","content":"test"}`
		resp, err := http.Post(addr+"/api/save", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("request %d: expected 201, got %d", i+1, resp.StatusCode)
		}
		resp.Body.Close()
	}

	body := `{"code":"rlt006","content":"test"}`
	resp, err := http.Post(addr+"/api/save", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429 on 6th request, got %d", resp.StatusCode)
	}
}

func TestGetRateLimit(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	saveBody := `{"code":"rlget1","content":"test"}`
	resp, _ := http.Post(addr+"/api/save", "application/json", strings.NewReader(saveBody))
	resp.Body.Close()

	for i := 0; i < 100; i++ {
		resp, err := http.Get(addr + "/rlget1")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, resp.StatusCode)
		}
		resp.Body.Close()
	}

	resp2, err := http.Get(addr + "/rlget1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429 on 101st request, got %d", resp2.StatusCode)
	}
}

func TestBodyTooLarge(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	bigBody := strings.Repeat("x", 5000)
	resp, err := http.Post(addr+"/api/save", "application/json", strings.NewReader(bigBody))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", resp.StatusCode)
	}
}

func TestCORSHeaders(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	resp, err := http.Get(addr + "/ZZZZZZ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got %q", resp.Header.Get("X-Content-Type-Options"))
	}
}

func TestUnknownRoute(t *testing.T) {
	addr, cleanup := newTestServer(t)
	defer cleanup()

	resp, err := http.Get(addr + "/api/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
