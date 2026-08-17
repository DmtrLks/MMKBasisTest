package team

import (
	"strings"
	"unicode/utf8"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) NormalizeAndValidateCreate(
	currentUserID int64,
	request *CreateRequest,
) error {
	if currentUserID <= 0 {
		return ErrInvalidUserID
	}

	request.Name = strings.TrimSpace(request.Name)

	switch {
	case request.Name == "":
		return ErrNameRequired

	case utf8.RuneCountInString(request.Name) > 255:
		return ErrNameTooLong

	default:
		return nil
	}
}

func (v *Validator) ValidateInvite(
	currentUserID int64,
	teamID int64,
	request InviteRequest,
) error {
	switch {
	case currentUserID <= 0:
		return ErrInvalidUserID

	case teamID <= 0:
		return ErrInvalidTeamID

	case request.UserID <= 0:
		return ErrInvalidMemberID

	case !request.Role.IsValid():
		return ErrInvalidRole

	case request.Role == RoleOwner:
		return ErrOwnerRoleDenied

	default:
		return nil
	}
}

func (v *Validator) ValidateUpdateMemberRole(
	currentUserID int64,
	teamID int64,
	memberUserID int64,
	request UpdateMemberRoleRequest,
) error {
	switch {
	case currentUserID <= 0:
		return ErrInvalidUserID

	case teamID <= 0:
		return ErrInvalidTeamID

	case memberUserID <= 0:
		return ErrInvalidMemberID

	case !request.Role.IsValid():
		return ErrInvalidRole

	case request.Role == RoleOwner:
		return ErrOwnerRoleDenied

	default:
		return nil
	}
}
