package comment

import "errors"

var (
	ErrInvalidUserID   = errors.New("invalid user ID")
	ErrInvalidTaskID   = errors.New("invalid task ID")
	ErrContentRequired = errors.New("comment content is required")
	ErrContentTooLong  = errors.New("comment content is too long")
	ErrTaskNotFound    = errors.New("task not found")
	ErrForbidden       = errors.New("user is not a team member")
)
