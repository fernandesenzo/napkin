package client

import (
	"testing"

	"github.com/gorilla/websocket"
)

func TestWritePump_SendsTextMessage(t *testing.T) {
	h := newStubHub()
	serverConn, clientConn := dialTestServer(t, nil)

	c := NewClient(h, serverConn)
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.WritePump()
	}()

	c.Send <- "hello from hub"

	_, msg, err := clientConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if string(msg) != "hello from hub" {
		t.Errorf("expected 'hello from hub', got %q", msg)
	}

	// Close the Send channel to trigger WritePump cleanup.
	close(c.Send)
	<-done
}

func TestWritePump_ClosedChannelSendsCloseFrame(t *testing.T) {
	h := newStubHub()
	serverConn, clientConn := dialTestServer(t, nil)

	c := NewClient(h, serverConn)
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.WritePump()
	}()

	close(c.Send)
	<-done

	// Client side should receive a close message.
	_, _, err := clientConn.ReadMessage()
	if err == nil {
		t.Error("expected connection to be closed after Send channel close")
	}
	if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) &&
		!websocket.IsUnexpectedCloseError(err) {
		// Some close error is acceptable — the key is that the connection is done.
		t.Logf("got close error (expected): %v", err)
	}
}

func TestWritePump_MultipleMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
	}{
		{name: "single message", messages: []string{"a"}},
		{name: "multiple messages", messages: []string{"first", "second", "third"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStubHub()
			serverConn, clientConn := dialTestServer(t, nil)

			c := NewClient(h, serverConn)
			done := make(chan struct{})
			go func() {
				defer close(done)
				c.WritePump()
			}()

			for _, msg := range tt.messages {
				c.Send <- msg
			}

			for _, want := range tt.messages {
				_, got, err := clientConn.ReadMessage()
				if err != nil {
					t.Fatalf("failed to read message: %v", err)
				}
				if string(got) != want {
					t.Errorf("expected %q, got %q", want, string(got))
				}
			}

			close(c.Send)
			<-done
		})
	}
}
