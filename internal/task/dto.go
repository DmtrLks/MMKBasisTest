package task

import "time"

type CreateRequest struct {
	TeamID      int64      `json:"team_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	AssigneeID  *int64     `json:"assignee_id"`
}

type UpdateRequest struct {
	Title       *string     `json:"title"`
	Description *string     `json:"description"`
	Status      *TaskStatus `json:"status"`
	AssigneeID  *int64      `json:"assignee_id"`
	Version     int64       `json:"version"`
}

type Response struct {
	ID          int64      `json:"id"`
	TeamID      int64      `json:"team_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	CreatedBy   int64      `json:"created_by"`
	AssigneeID  *int64     `json:"assignee_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	Version     int64      `json:"version"`
}

func responseFromTask(task Task) Response {
	return Response{
		ID:          task.ID,
		TeamID:      task.TeamID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedBy:   task.CreatedBy,
		AssigneeID:  task.AssigneeID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		ClosedAt:    task.ClosedAt,
		Version:     task.Version,
	}
}
