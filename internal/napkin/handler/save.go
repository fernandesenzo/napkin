package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
		return
	}
	var req saveNapkinRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	npk, err := h.svc.Save(r.Context(), req.Code, req.Content)
	if err != nil {
		handleSaveError(w, err, r.Context())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := saveNapkinResponse{
		Content: npk.Text,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.ErrorContext(r.Context(), "handler.Create: failed to encode response", "err", err.Error())
	}

}

func handleSaveError(w http.ResponseWriter, err error, ctx context.Context) {
	switch {
	case errors.Is(err, napkin.ErrContentTooLong):
		http.Error(w, "content too long", http.StatusUnprocessableEntity)
	case errors.Is(err, napkin.ErrInvalidCode):
		http.Error(w, "invalid code", http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
		slog.ErrorContext(ctx, "handler.Save: unknown error on save request", "err", err)
	}
}
