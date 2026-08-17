package history

import "errors"

var (
	ErrInvalidUserID = errors.New("invalid user ID")
	ErrInvalidTaskID = errors.New("invalid task ID")
	ErrTaskNotFound  = errors.New("task not found")
	ErrForbidden     = errors.New("user is not a team member")
)
