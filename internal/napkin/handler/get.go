package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	npk, err := h.svc.Get(r.Context(), code)
	if err != nil {
		handleGetError(w, r.Context(), err)
		return
	}
	var resp getNapkinResponse
	resp.Code = npk.Code
	resp.Content = npk.Text

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.ErrorContext(r.Context(), "handler.Get: failed to encode response", "err", err.Error())
	}

}

func handleGetError(w http.ResponseWriter, ctx context.Context, err error) {
	switch {
	case errors.Is(err, napkin.ErrInvalidCode):
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	case errors.Is(err, napkin.ErrNapkinDoesNotExist):
		http.Error(w, "napkin not found", http.StatusNotFound)
		return
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
		slog.ErrorContext(ctx, "handler.Get: unknown error obtaining napkin", "err", err)
		return
	}
}
