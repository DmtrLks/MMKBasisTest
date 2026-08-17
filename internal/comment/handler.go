package comment

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	taskID, err := parseTaskID(r)
	if err != nil {
		h.handleError(w, err)

		return
	}

	var request CreateRequest
	if err := httpg.DecodeJSON(w, r, &request); err != nil {
		httpg.WriteInvalidRequest(w, err)

		return
	}

	response, err := h.service.Create(r.Context(), currentUserID, taskID, request)
	if err != nil {
		h.handleError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	taskID, err := parseTaskID(r)
	if err != nil {
		h.handleError(w, err)

		return
	}

	response, err := h.service.List(r.Context(), currentUserID, taskID)
	if err != nil {
		h.handleError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusOK, response)
}

func parseTaskID(r *http.Request) (int64, error) {
	taskID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || taskID <= 0 {
		return 0, ErrInvalidTaskID
	}

	return taskID, nil
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		httpg.WriteUnauthorized(w)

	case errors.Is(err, ErrInvalidTaskID),
		errors.Is(err, ErrContentRequired),
		errors.Is(err, ErrContentTooLong):
		httpg.WriteValidationError(w, err)

	case errors.Is(err, ErrTaskNotFound):
		httpg.WriteError(w, http.StatusNotFound, "task_not_found", "task not found")

	case errors.Is(err, ErrForbidden):
		httpg.WriteForbidden(w, "user is not a team member")

	default:
		httpg.WriteInternalError(w, "comment request", err)
	}
}
