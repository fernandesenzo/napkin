package handler

import (
	"log/slog"
	"net/http"

	"github.com/fernandesenzo/napkin/internal/client"
	"github.com/fernandesenzo/napkin/internal/napkin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
	CheckOrigin: func(r *http.Request) bool {
		return true //check for domain origin on production
	},
}

func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}
	if err := napkin.ValidateCode(code); err != nil {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.ErrorContext(r.Context(), "handler.WebSocket: failed to upgrade connection", "err", err)
		return
	}

	roomHub := h.hubManager.GetOrCreateRoom(code)

	clientObj := client.NewClient(roomHub, conn)

	if !roomHub.Join(clientObj) {
		roomHub = h.hubManager.GetOrCreateRoom(code)
		clientObj = client.NewClient(roomHub, conn)
		if !roomHub.Join(clientObj) {
			conn.Close()
			return
		}
	}

	go clientObj.WritePump()
	go clientObj.ReadPump()
}
