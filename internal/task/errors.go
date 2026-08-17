package task

import "errors"

var (
	ErrInvalidUserID   = errors.New("invalid user ID")
	ErrInvalidTaskID   = errors.New("invalid task ID")
	ErrInvalidTeamID   = errors.New("invalid team ID")
	ErrTitleRequired   = errors.New("task title is required")
	ErrTitleTooLong    = errors.New("task title is too long")
	ErrInvalidStatus   = errors.New("invalid task status")
	ErrInvalidAssignee = errors.New("invalid assignee ID")
	ErrInvalidLimit    = errors.New("limit must be between 1 and 100")
	ErrInvalidOffset   = errors.New("offset cannot be negative")
	ErrInvalidVersion  = errors.New("task version must be positive")
	ErrNoUpdateFields  = errors.New("at least one task field must be provided")
	ErrNoChanges       = errors.New("task fields have not changed")
	ErrTaskNotFound    = errors.New("task not found")
	ErrVersionConflict = errors.New("task was changed by another request")
	ErrNotTeamMember   = errors.New("user is not a team member")
	ErrUpdateForbidden = errors.New("user cannot update this task")
	ErrAssigneeNotTeam = errors.New("assignee is not a team member")
)
