package task

import "time"

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusTodo,
		TaskStatusInProgress,
		TaskStatusDone:
		return true

	default:
		return false
	}
}

type Task struct {
	ID          int64
	TeamID      int64
	Title       string
	Description string
	Status      TaskStatus
	CreatedBy   int64
	AssigneeID  *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ClosedAt    *time.Time
	Version     int64
}

type TaskParams struct {
	TeamID     int64
	Status     *TaskStatus
	AssigneeID *int64
	Limit      int
	Offset     int
}
