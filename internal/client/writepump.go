package client

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			slog.Debug("client.WritePump: failed to close connection on defer", "err", err)
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				slog.Debug("client.WritePump: failed to set write deadline for message", "err", err)
				return
			}

			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					slog.Debug("client.WritePump: failed to send close message", "err", err)
				}
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				slog.Debug("client.WritePump: failed to write text message", "err", err)
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				slog.Debug("client.WritePump: failed to set write deadline for ping", "err", err)
				return
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Debug("client.WritePump: failed to send ping message", "err", err)
				return
			}
		}
	}
}
