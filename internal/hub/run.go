package hub

import (
	"context"
	"log/slog"
)

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			slog.Info("hub.Run: a new client entered the room", "code", h.Code)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				slog.Info("hub.Run: a client left the room", "code", h.Code)
				if len(h.clients) == 0 {
					if h.OnEmpty != nil {
						h.OnEmpty()
					}
					return
				}
			}
		case message := <-h.broadcast:
			go func(msg string) {
				if _, err := h.svc.Save(context.Background(), h.Code, msg); err != nil {
					slog.Error("hub.Run: failed persisting state", "err", err, "code", h.Code)
				}
			}(message)
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}

	}
}
