package auth

import (
	"errors"
	"mmktestbasisByDGanichev/internal/httpg"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request LoginRequest

	if err := httpg.DecodeJSON(w, r, &request); err != nil {
		httpg.WriteInvalidRequest(w, err)

		return
	}

	response, err := h.service.Login(r.Context(), request)
	if err != nil {
		h.handleLoginError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleLoginError(
	w http.ResponseWriter,
	err error,
) {
	if errors.Is(err, ErrInvalidCredentials) {
		httpg.WriteError(
			w,
			http.StatusUnauthorized,
			"invalid_credentials",
			"invalid email or password",
		)

		return
	}

	httpg.WriteInternalError(w, "login user", err)
}
