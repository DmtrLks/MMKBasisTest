package mysql

import (
	"context"
	"fmt"

	"mmktestbasisByDGanichev/internal/comment"
	"mmktestbasisByDGanichev/internal/database"
)

type CommentRepository struct{}

func NewCommentRepository() *CommentRepository {
	return &CommentRepository{}
}

func (r *CommentRepository) Create(
	ctx context.Context,
	db database.DBTX,
	comment comment.Comment,
) (comment.Comment, error) {
	const query = `
		INSERT INTO task_comments (
			task_id,
			user_id,
			content,
			created_at
		)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.ExecContext(
		ctx,
		query,
		comment.TaskID,
		comment.UserID,
		comment.Content,
		comment.CreatedAt,
	)
	if err != nil {
		return comment, fmt.Errorf("insert task comment: %w", err)
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return comment, fmt.Errorf("get inserted task comment ID: %w", err)
	}

	comment.ID = commentID

	return comment, nil
}

func (r *CommentRepository) ListByTaskID(
	ctx context.Context,
	db database.DBTX,
	taskID int64,
) ([]comment.Comment, error) {
	const query = `
		SELECT
			id,
			task_id,
			user_id,
			content,
			created_at
		FROM task_comments
		WHERE task_id = ?
		ORDER BY created_at ASC, id ASC
	`

	rows, err := db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("list task comments: %w", err)
	}
	defer rows.Close()

	comments := make([]comment.Comment, 0)
	for rows.Next() {
		var item comment.Comment

		if err := rows.Scan(
			&item.ID,
			&item.TaskID,
			&item.UserID,
			&item.Content,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task comment: %w", err)
		}

		comments = append(comments, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate task comments: %w", err)
	}

	return comments, nil
}
