package stats

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

func (h *Handler) Report(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	teamID, err := strconv.ParseInt(r.PathValue("team_id"), 10, 64)
	if err != nil || teamID <= 0 {
		h.handleError(w, ErrInvalidTeamID)

		return
	}

	report, err := h.service.Report(r.Context(), currentUserID, teamID)
	if err != nil {
		h.handleError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusOK, report)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		httpg.WriteUnauthorized(w)

	case errors.Is(err, ErrInvalidTeamID):
		httpg.WriteValidationError(w, err)

	case errors.Is(err, ErrForbidden):
		httpg.WriteForbidden(w, "team statistics are available only to owner or admin")

	default:
		httpg.WriteInternalError(w, "get team statistics", err)
	}
}
