package client

import (
	"log/slog"
	"time"

	"github.com/fernandesenzo/napkin/internal/napkin"
	"github.com/gorilla/websocket"
)

func (c *Client) ReadPump() {
	defer func() {
		c.hub.UnregisterChan() <- c
		if err := c.conn.Close(); err != nil {
			slog.Debug("client.ReadPump: failed to close connection on defer", "err", err)
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		slog.Debug("client.ReadPump: failed to set initial read deadline", "err", err)
		return
	}

	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			slog.Debug("client.ReadPump: failed to update read deadline in pong handler", "err", err)
			return err
		}
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("client.ReadPump: unexpected websocket closure", "err", err)
			}
			break
		}

		message := string(messageBytes)

		if err := napkin.ValidateContent(message); err != nil {
			slog.Warn("client.ReadPump: invalid content size", "err", err, "room", c.hub.GetCode())
			continue
		}
		if !c.limiter.Allow() {
			slog.Warn("client.ReadPump: rate limit exceeded in memory for client", "room", c.hub.GetCode())
			continue
		}
		c.hub.BroadcastChan() <- message
	}
}
