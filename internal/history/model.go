package history

import "time"

type FieldChange struct {
	Old any `json:"old"`
	New any `json:"new"`
}

type Changes map[string]FieldChange

type Entry struct {
	ID        int64
	TaskID    int64
	ChangedBy int64
	Changes   Changes
	CreatedAt time.Time
}
