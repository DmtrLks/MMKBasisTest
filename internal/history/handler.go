package history

import (
	"errors"
	"net/http"
	"strconv"

	"mmktestbasisByDGanichev/internal/httpg"
	"mmktestbasisByDGanichev/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	taskID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || taskID <= 0 {
		h.handleError(w, ErrInvalidTaskID)

		return
	}

	response, err := h.service.List(r.Context(), currentUserID, taskID)
	if err != nil {
		h.handleError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		httpg.WriteUnauthorized(w)

	case errors.Is(err, ErrInvalidTaskID):
		httpg.WriteValidationError(w, err)

	case errors.Is(err, ErrTaskNotFound):
		httpg.WriteError(w, http.StatusNotFound, "task_not_found", "task not found")

	case errors.Is(err, ErrForbidden):
		httpg.WriteForbidden(w, "user is not a team member")

	default:
		httpg.WriteInternalError(w, "list task history", err)
	}
}
