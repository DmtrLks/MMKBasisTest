package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"mmktestbasisByDGanichev/internal/database"
	"mmktestbasisByDGanichev/internal/history"
)

type HistoryRepository struct{}

func NewHistoryRepository() *HistoryRepository {
	return &HistoryRepository{}
}

func (r *HistoryRepository) Append(
	ctx context.Context,
	db database.DBTX,
	entry history.Entry,
) (history.Entry, error) {
	changes, err := json.Marshal(entry.Changes)
	if err != nil {
		return history.Entry{}, fmt.Errorf("marshal task history changes: %w", err)
	}

	const query = `
		INSERT INTO task_history (
			task_id,
			changed_by,
			changes,
			created_at
		)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.ExecContext(
		ctx,
		query,
		entry.TaskID,
		entry.ChangedBy,
		string(changes),
		entry.CreatedAt,
	)
	if err != nil {
		return history.Entry{}, fmt.Errorf("insert task history: %w", err)
	}

	entryID, err := result.LastInsertId()
	if err != nil {
		return history.Entry{}, fmt.Errorf("get inserted task history ID: %w", err)
	}

	entry.ID = entryID

	return entry, nil
}

func (r *HistoryRepository) TeamIDByTaskID(
	ctx context.Context,
	db database.DBTX,
	taskID int64,
) (int64, error) {
	const query = `
		SELECT team_id
		FROM tasks
		WHERE id = ?
	`

	var teamID int64

	err := db.QueryRowContext(ctx, query, taskID).Scan(&teamID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, history.ErrTaskNotFound
	}

	if err != nil {
		return 0, fmt.Errorf("find task team ID: %w", err)
	}

	return teamID, nil
}

func (r *HistoryRepository) ListByTaskID(
	ctx context.Context,
	db database.DBTX,
	taskID int64,
) ([]history.Entry, error) {
	const query = `
		SELECT
			id,
			task_id,
			changed_by,
			changes,
			created_at
		FROM task_history
		WHERE task_id = ?
		ORDER BY created_at ASC, id ASC
	`

	rows, err := db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("list task history: %w", err)
	}
	defer rows.Close()

	entries := make([]history.Entry, 0)
	for rows.Next() {
		var entry history.Entry
		var changes []byte

		if err := rows.Scan(
			&entry.ID,
			&entry.TaskID,
			&entry.ChangedBy,
			&changes,
			&entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task history: %w", err)
		}

		if err := json.Unmarshal(changes, &entry.Changes); err != nil {
			return nil, fmt.Errorf("unmarshal task history changes: %w", err)
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate task history: %w", err)
	}

	return entries, nil
}
