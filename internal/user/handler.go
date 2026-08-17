package user

import (
	"errors"
	"net/http"

	httpg "mmktestbasisByDGanichev/internal/httpg"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest

	if err := httpg.DecodeJSON(w, r, &request); err != nil {
		httpg.WriteInvalidRequest(w, err)

		return
	}

	response, err := h.service.Register(r.Context(), request)
	if err != nil {
		h.handleRegisterError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) handleRegisterError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrEmailAlreadyExists):
		httpg.WriteError(
			w,
			http.StatusConflict,
			"email_already_exists",
			"user with this email already exists",
		)

	case errors.Is(err, ErrEmailRequired),
		errors.Is(err, ErrInvalidEmail),
		errors.Is(err, ErrEmailTooLong),
		errors.Is(err, ErrNameRequired),
		errors.Is(err, ErrNameTooLong),
		errors.Is(err, ErrPasswordTooShort),
		errors.Is(err, ErrPasswordTooLong):
		httpg.WriteValidationError(w, err)

	default:
		httpg.WriteInternalError(w, "register user", err)
	}
}
