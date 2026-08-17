package stats

import "errors"

var (
	ErrInvalidUserID = errors.New("invalid user ID")
	ErrInvalidTeamID = errors.New("invalid team ID")
	ErrForbidden     = errors.New("operation is forbidden")
)
