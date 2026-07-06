package client

import (
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestReadPump_ValidMessageIsBroadcast(t *testing.T) {
	h := newStubHub()
	serverConn, clientConn := dialTestServer(t, nil)

	c := NewClient(h, serverConn)
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.ReadPump()
	}()

	msg := "hello world"
	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	got := <-h.broadcast
	if got != msg {
		t.Errorf("expected broadcast %q, got %q", msg, got)
	}

	// Trigger cleanup: close client side so ReadPump exits.
	clientConn.Close()
	<-done

	unregistered := <-h.unregister
	if unregistered != c {
		t.Error("expected client to unregister itself on exit")
	}
}

func TestReadPump_ContentTooLongIsDropped(t *testing.T) {
	h := newStubHub()
	serverConn, clientConn := dialTestServer(t, nil)

	c := NewClient(h, serverConn)
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.ReadPump()
	}()

	// Send a message that exceeds MaxContentLength (200 bytes).
	tooLong := strings.Repeat("x", 201)
	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(tooLong)); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	// Send a valid message right after so we can verify ordering.
	valid := "valid"
	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(valid)); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	got := <-h.broadcast
	if got != valid {
		t.Errorf("expected only valid message %q to be broadcast, got %q", valid, got)
	}

	clientConn.Close()
	<-done
	<-h.unregister
}

func TestReadPump_UnregistersOnClose(t *testing.T) {
	h := newStubHub()
	serverConn, clientConn := dialTestServer(t, nil)

	c := NewClient(h, serverConn)
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.ReadPump()
	}()

	clientConn.Close()
	<-done

	select {
	case unregistered := <-h.unregister:
		if unregistered != c {
			t.Error("wrong client was unregistered")
		}
	default:
		t.Error("expected client to be sent to unregister channel")
	}
}
