package task

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

	var request CreateRequest

	if err := httpg.DecodeJSON(w, r, &request); err != nil {
		httpg.WriteInvalidRequest(w, err)

		return
	}

	response, err := h.service.Create(r.Context(), currentUserID, request)
	if err != nil {
		h.handleCreateError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	params, err := parseTaskParams(r)
	if err != nil {
		h.handleListError(w, err)

		return
	}

	response, err := h.service.List(r.Context(), currentUserID, params)
	if err != nil {
		h.handleListError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	taskID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.handleUpdateError(w, ErrInvalidTaskID)

		return
	}

	var request UpdateRequest

	if err := httpg.DecodeJSON(w, r, &request); err != nil {
		httpg.WriteInvalidRequest(w, err)

		return
	}

	response, err := h.service.Update(r.Context(), currentUserID, taskID, request)
	if err != nil {
		h.handleUpdateError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusOK, response)
}

func parseTaskParams(r *http.Request) (TaskParams, error) {
	query := r.URL.Query()

	teamID, err := strconv.ParseInt(query.Get("team_id"), 10, 64)
	if err != nil {
		return TaskParams{}, ErrInvalidTeamID
	}

	params := TaskParams{
		TeamID: teamID,
		Limit:  20,
	}

	if rawStatus := query.Get("status"); rawStatus != "" {
		status := TaskStatus(rawStatus)
		params.Status = &status
	}

	if rawAssigneeID := query.Get("assignee_id"); rawAssigneeID != "" {
		assigneeID, err := strconv.ParseInt(rawAssigneeID, 10, 64)
		if err != nil {
			return TaskParams{}, ErrInvalidAssignee
		}

		params.AssigneeID = &assigneeID
	}

	if rawLimit := query.Get("limit"); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil {
			return TaskParams{}, ErrInvalidLimit
		}

		params.Limit = limit
	}

	if rawOffset := query.Get("offset"); rawOffset != "" {
		offset, err := strconv.Atoi(rawOffset)
		if err != nil {
			return TaskParams{}, ErrInvalidOffset
		}

		params.Offset = offset
	}

	return params, nil
}

func (h *Handler) handleCreateError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		httpg.WriteUnauthorized(w)

	case errors.Is(err, ErrInvalidTeamID),
		errors.Is(err, ErrTitleRequired),
		errors.Is(err, ErrTitleTooLong),
		errors.Is(err, ErrInvalidStatus),
		errors.Is(err, ErrInvalidAssignee):
		httpg.WriteValidationError(w, err)

	case errors.Is(err, ErrNotTeamMember):
		httpg.WriteForbidden(w, "user is not a team member")

	case errors.Is(err, ErrAssigneeNotTeam):
		httpg.WriteError(
			w,
			http.StatusBadRequest,
			"assignee_not_team_member",
			"assignee is not a team member",
		)

	default:
		httpg.WriteInternalError(w, "create task", err)
	}
}

func (h *Handler) handleListError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		httpg.WriteUnauthorized(w)

	case errors.Is(err, ErrInvalidTeamID),
		errors.Is(err, ErrInvalidStatus),
		errors.Is(err, ErrInvalidAssignee),
		errors.Is(err, ErrInvalidLimit),
		errors.Is(err, ErrInvalidOffset):
		httpg.WriteValidationError(w, err)

	case errors.Is(err, ErrNotTeamMember):
		httpg.WriteForbidden(w, "user is not a team member")

	default:
		httpg.WriteInternalError(w, "list tasks", err)
	}
}

func (h *Handler) handleUpdateError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		httpg.WriteUnauthorized(w)

	case errors.Is(err, ErrInvalidTaskID),
		errors.Is(err, ErrInvalidVersion),
		errors.Is(err, ErrNoUpdateFields),
		errors.Is(err, ErrNoChanges),
		errors.Is(err, ErrTitleRequired),
		errors.Is(err, ErrTitleTooLong),
		errors.Is(err, ErrInvalidStatus),
		errors.Is(err, ErrInvalidAssignee):
		httpg.WriteValidationError(w, err)

	case errors.Is(err, ErrTaskNotFound):
		httpg.WriteError(w, http.StatusNotFound, "task_not_found", "task not found")

	case errors.Is(err, ErrNotTeamMember),
		errors.Is(err, ErrUpdateForbidden):
		httpg.WriteForbidden(w, "user is not a team member")

	case errors.Is(err, ErrAssigneeNotTeam):
		httpg.WriteError(
			w,
			http.StatusBadRequest,
			"assignee_not_team_member",
			"assignee is not a team member",
		)

	case errors.Is(err, ErrVersionConflict):
		httpg.WriteError(
			w,
			http.StatusConflict,
			"version_conflict",
			"task was changed by another request",
		)

	default:
		httpg.WriteInternalError(w, "update task", err)
	}
}
