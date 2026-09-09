package handler

import (
	"log/slog"
	"net/http"

	"github.com/fernandesenzo/napkin/internal/client"
	"github.com/fernandesenzo/napkin/internal/napkin"
)

func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}
	if err := napkin.ValidateCode(code, h.config.CodeLength); err != nil {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.ErrorContext(r.Context(), "handler.WebSocket: failed to upgrade connection", "err", err)
		return
	}

	roomHub := h.hubManager.GetOrCreateRoom(code)

	clientConfig := client.Config{MaxContentLength: h.config.MaxContentLength}
	clientObj := client.NewClient(roomHub, conn, clientConfig)

	if !roomHub.Join(clientObj) {
		roomHub = h.hubManager.GetOrCreateRoom(code)
		clientObj = client.NewClient(roomHub, conn, clientConfig)
		if !roomHub.Join(clientObj) {
			conn.Close()
			return
		}
	}

	go clientObj.WritePump()
	go clientObj.ReadPump()
}
