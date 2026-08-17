package team

import (
	"errors"
	"mmktestbasisByDGanichev/internal/httpg"
	"mmktestbasisByDGanichev/internal/middleware"
	"net/http"
	"strconv"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	var request CreateRequest

	if err := httpg.DecodeJSON(w, r, &request); err != nil {
		httpg.WriteInvalidRequest(w, err)

		return
	}

	response, err := h.service.Create(r.Context(), userID, request)
	if err != nil {
		h.handleCreateError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) handleCreateError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		httpg.WriteUnauthorized(w)

	case errors.Is(err, ErrNameRequired),
		errors.Is(err, ErrNameTooLong):
		httpg.WriteValidationError(w, err)

	default:
		httpg.WriteInternalError(w, "create team", err)
	}
}

func (h *Handler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	response, err := h.service.List(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrInvalidUserID) {
			httpg.WriteUnauthorized(w)

			return
		}

		httpg.WriteInternalError(w, "list teams", err)

		return
	}

	httpg.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Invite(
	w http.ResponseWriter,
	r *http.Request,
) {
	currentUserID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	teamID, ok := parsePathID(w, r, "id", "invalid_team_id", "team ID must be a positive integer")
	if !ok {
		return
	}

	var request InviteRequest

	if err := httpg.DecodeJSON(w, r, &request); err != nil {
		httpg.WriteInvalidRequest(w, err)

		return
	}

	response, err := h.service.Invite(r.Context(), currentUserID, teamID, request)
	if err != nil {
		h.handleInviteError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) UpdateMemberRole(
	w http.ResponseWriter,
	r *http.Request,
) {
	currentUserID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	teamID, ok := parsePathID(w, r, "id", "invalid_team_id", "team ID must be a positive integer")
	if !ok {
		return
	}

	memberUserID, ok := parsePathID(
		w,
		r,
		"user_id",
		"invalid_member_id",
		"member user ID must be a positive integer",
	)
	if !ok {
		return
	}

	var request UpdateMemberRoleRequest

	if err := httpg.DecodeJSON(w, r, &request); err != nil {
		httpg.WriteInvalidRequest(w, err)

		return
	}

	response, err := h.service.UpdateMemberRole(
		r.Context(),
		currentUserID,
		teamID,
		memberUserID,
		request,
	)
	if err != nil {
		h.handleUpdateMemberRoleError(w, err)

		return
	}

	httpg.WriteJSON(w, http.StatusOK, response)
}

func parsePathID(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	errorCode string,
	message string,
) (int64, bool) {
	value, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || value <= 0 {
		httpg.WriteError(w, http.StatusBadRequest, errorCode, message)

		return 0, false
	}

	return value, true
}

func (h *Handler) handleUpdateMemberRoleError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		httpg.WriteUnauthorized(w)

	case errors.Is(err, ErrInvalidTeamID),
		errors.Is(err, ErrInvalidMemberID),
		errors.Is(err, ErrInvalidRole),
		errors.Is(err, ErrOwnerRoleDenied):
		httpg.WriteValidationError(w, err)

	case errors.Is(err, ErrForbidden):
		httpg.WriteForbidden(w, "only the team owner can change member roles")

	case errors.Is(err, ErrOwnerRoleImmutable):
		httpg.WriteError(
			w,
			http.StatusForbidden,
			"owner_role_immutable",
			"team owner role cannot be changed",
		)

	case errors.Is(err, ErrMemberNotFound):
		httpg.WriteError(w, http.StatusNotFound, "member_not_found", "team member not found")

	default:
		httpg.WriteInternalError(w, "update team member role", err)
	}
}

func (h *Handler) handleInviteError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		httpg.WriteUnauthorized(w)

	case errors.Is(err, ErrInvalidTeamID),
		errors.Is(err, ErrInvalidMemberID),
		errors.Is(err, ErrInvalidRole),
		errors.Is(err, ErrOwnerRoleDenied):
		httpg.WriteValidationError(w, err)

	case errors.Is(err, ErrForbidden):
		httpg.WriteForbidden(w, "insufficient team permissions")

	case errors.Is(err, ErrInvitedUserAbsent):
		httpg.WriteError(w, http.StatusNotFound, "user_not_found", "invited user not found")

	case errors.Is(err, ErrAlreadyMember):
		httpg.WriteError(
			w,
			http.StatusConflict,
			"already_team_member",
			"user is already a team member",
		)

	default:
		httpg.WriteInternalError(w, "invite team member", err)
	}
}
