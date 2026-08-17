package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"mmktestbasisByDGanichev/internal/database"
	"mmktestbasisByDGanichev/internal/task"
)

type TaskRepository struct{}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{}
}

func (r *TaskRepository) Create(
	ctx context.Context,
	db database.DBTX,
	t task.Task,
) (task.Task, error) {
	const query = `
		INSERT INTO tasks (
			team_id,
			title,
			description,
			status,
			created_by,
			assignee_id,
			created_at,
			updated_at,
			closed_at,
			version
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.ExecContext(
		ctx,
		query,
		t.TeamID,
		t.Title,
		t.Description,
		t.Status,
		t.CreatedBy,
		t.AssigneeID,
		t.CreatedAt,
		t.UpdatedAt,
		t.ClosedAt,
		t.Version,
	)
	if err != nil {
		return task.Task{}, fmt.Errorf("insert task: %w", err)
	}

	taskID, err := result.LastInsertId()
	if err != nil {
		return task.Task{}, fmt.Errorf("get inserted task ID: %w", err)
	}

	t.ID = taskID

	return t, nil
}

func (r *TaskRepository) List(
	ctx context.Context,
	db database.DBTX,
	params task.TaskParams,
) ([]task.Task, error) {
	var query strings.Builder

	query.WriteString(`
		SELECT
			id,
			team_id,
			title,
			COALESCE(description, ''),
			status,
			created_by,
			assignee_id,
			created_at,
			updated_at,
			closed_at,
			version
		FROM tasks
		WHERE team_id = ?
	`)

	args := []any{params.TeamID}

	if params.Status != nil {
		query.WriteString(" AND status = ?")
		args = append(args, *params.Status)
	}

	if params.AssigneeID != nil {
		query.WriteString(" AND assignee_id = ?")
		args = append(args, *params.AssigneeID)
	}

	query.WriteString(" ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?")
	args = append(args, params.Limit, params.Offset)

	rows, err := db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	result := make([]task.Task, 0)

	for rows.Next() {
		var item task.Task
		var status string
		var assigneeID sql.NullInt64
		var closedAt sql.NullTime

		if err := rows.Scan(
			&item.ID,
			&item.TeamID,
			&item.Title,
			&item.Description,
			&status,
			&item.CreatedBy,
			&assigneeID,
			&item.CreatedAt,
			&item.UpdatedAt,
			&closedAt,
			&item.Version,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		item.Status = task.TaskStatus(status)
		if !item.Status.IsValid() {
			return nil, fmt.Errorf("invalid task status %q in database", status)
		}

		if assigneeID.Valid {
			value := assigneeID.Int64
			item.AssigneeID = &value
		}

		if closedAt.Valid {
			value := closedAt.Time
			item.ClosedAt = &value
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return result, nil
}

func (r *TaskRepository) FindByID(
	ctx context.Context,
	db database.DBTX,
	taskID int64,
) (task.Task, error) {
	const query = `
		SELECT
			id,
			team_id,
			title,
			COALESCE(description, ''),
			status,
			created_by,
			assignee_id,
			created_at,
			updated_at,
			closed_at,
			version
		FROM tasks
		WHERE id = ?
	`

	var item task.Task
	var status string
	var assigneeID sql.NullInt64
	var closedAt sql.NullTime

	err := db.QueryRowContext(ctx, query, taskID).Scan(
		&item.ID,
		&item.TeamID,
		&item.Title,
		&item.Description,
		&status,
		&item.CreatedBy,
		&assigneeID,
		&item.CreatedAt,
		&item.UpdatedAt,
		&closedAt,
		&item.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return task.Task{}, task.ErrTaskNotFound
	}

	if err != nil {
		return task.Task{}, fmt.Errorf("find task by ID: %w", err)
	}

	item.Status = task.TaskStatus(status)
	if !item.Status.IsValid() {
		return task.Task{}, fmt.Errorf("invalid task status %q in database", status)
	}

	if assigneeID.Valid {
		value := assigneeID.Int64
		item.AssigneeID = &value
	}

	if closedAt.Valid {
		value := closedAt.Time
		item.ClosedAt = &value
	}

	return item, nil
}

func (r *TaskRepository) Update(
	ctx context.Context,
	db database.DBTX,
	t task.Task,
	expectedVersion int64,
) (task.Task, error) {
	const query = `
		UPDATE tasks
		SET
			title = ?,
			description = ?,
			status = ?,
			assignee_id = ?,
			updated_at = ?,
			closed_at = ?,
			version = version + 1
		WHERE id = ?
		  AND version = ?
	`

	result, err := db.ExecContext(
		ctx,
		query,
		t.Title,
		t.Description,
		t.Status,
		t.AssigneeID,
		t.UpdatedAt,
		t.ClosedAt,
		t.ID,
		expectedVersion,
	)
	if err != nil {
		return task.Task{}, fmt.Errorf("update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return task.Task{}, fmt.Errorf("get updated task rows count: %w", err)
	}

	if rowsAffected == 0 {
		return task.Task{}, task.ErrVersionConflict
	}

	return t, nil
}
