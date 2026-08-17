package history

import "time"

type Response struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	ChangedBy int64     `json:"changed_by"`
	Changes   Changes   `json:"changes"`
	CreatedAt time.Time `json:"created_at"`
}

func responseFromEntry(entry Entry) Response {
	return Response{
		ID:        entry.ID,
		TaskID:    entry.TaskID,
		ChangedBy: entry.ChangedBy,
		Changes:   entry.Changes,
		CreatedAt: entry.CreatedAt,
	}
}
