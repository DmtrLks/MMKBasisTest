package team

import "errors"

var (
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrInvalidTeamID      = errors.New("invalid team ID")
	ErrInvalidMemberID    = errors.New("invalid member user ID")
	ErrNameRequired       = errors.New("team name is required")
	ErrNameTooLong        = errors.New("team name is too long")
	ErrInvalidRole        = errors.New("invalid team role")
	ErrOwnerRoleDenied    = errors.New("owner role cannot be assigned")
	ErrOwnerRoleImmutable = errors.New("team owner role cannot be changed")
	ErrForbidden          = errors.New("operation is forbidden")
	ErrMemberNotFound     = errors.New("team member not found")
	ErrAlreadyMember      = errors.New("user is already a team member")
	ErrInvitedUserAbsent  = errors.New("invited user not found")
)
