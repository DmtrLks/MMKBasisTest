package stats

type TaskCounts struct {
	Todo       int64 `json:"todo"`
	InProgress int64 `json:"in_progress"`
	Done       int64 `json:"done"`
}

type TopAssignee struct {
	UserID      int64  `json:"user_id"`
	Name        string `json:"name"`
	ClosedTasks int64  `json:"closed_tasks"`
}

type Report struct {
	TeamID                  int64         `json:"team_id"`
	TasksByStatus           TaskCounts    `json:"tasks_by_status"`
	TopAssignees            []TopAssignee `json:"top_assignees"`
	AverageCloseTimeSeconds *float64      `json:"average_close_time_seconds"`
	CommentsCount           int64         `json:"comments_count"`
}
