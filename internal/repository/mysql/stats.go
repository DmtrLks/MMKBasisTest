package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"mmktestbasisByDGanichev/internal/database"
	"mmktestbasisByDGanichev/internal/stats"
	"mmktestbasisByDGanichev/internal/task"
)

type StatsRepository struct{}

func NewStatsRepository() *StatsRepository {
	return &StatsRepository{}
}

func (r *StatsRepository) Report(
	ctx context.Context,
	db database.DBTX,
	teamID int64,
) (stats.Report, error) {
	const query = `
		SELECT
			(
				SELECT COUNT(CASE WHEN status = ? THEN 1 END)
				FROM tasks
				WHERE team_id = ?
			) AS todo_count,
			(
				SELECT COUNT(CASE WHEN status = ? THEN 1 END)
				FROM tasks
				WHERE team_id = ?
			) AS in_progress_count,
			(
				SELECT COUNT(CASE WHEN status = ? THEN 1 END)
				FROM tasks
				WHERE team_id = ?
			) AS done_count,
			(
				SELECT AVG(
					CASE
						WHEN status = ? AND closed_at IS NOT NULL
						THEN TIMESTAMPDIFF(SECOND, created_at, closed_at)
					END
				)
				FROM tasks
				WHERE team_id = ?
			) AS average_close_time_seconds,
			(
				SELECT COUNT(tc.id)
				FROM tasks AS t
				LEFT JOIN task_comments AS tc ON tc.task_id = t.id
				WHERE t.team_id = ?
			) AS comments_count,
			ta.user_id,
			ta.name,
			ta.closed_tasks
		FROM (SELECT 1) AS base
		LEFT JOIN (
			SELECT
				t.assignee_id AS user_id,
				u.name,
				COUNT(*) AS closed_tasks
			FROM tasks AS t
			INNER JOIN users AS u ON u.id = t.assignee_id
			WHERE t.team_id = ?
			  AND t.status = ?
			  AND t.closed_at >= UTC_TIMESTAMP() - INTERVAL 30 DAY
			GROUP BY t.assignee_id, u.name
			ORDER BY closed_tasks DESC, user_id ASC
			LIMIT 3
		) AS ta ON TRUE
		ORDER BY ta.closed_tasks DESC, ta.user_id ASC
	`

	rows, err := db.QueryContext(
		ctx,
		query,
		task.TaskStatusTodo,
		teamID,
		task.TaskStatusInProgress,
		teamID,
		task.TaskStatusDone,
		teamID,
		task.TaskStatusDone,
		teamID,
		teamID,
		teamID,
		task.TaskStatusDone,
	)
	if err != nil {
		return stats.Report{}, fmt.Errorf("query team statistics: %w", err)
	}
	defer rows.Close()

	report := stats.Report{
		TeamID:       teamID,
		TopAssignees: make([]stats.TopAssignee, 0, 3),
	}

	for rows.Next() {
		var averageCloseTime sql.NullFloat64
		var userID sql.NullInt64
		var name sql.NullString
		var closedTasks sql.NullInt64

		if err := rows.Scan(
			&report.TasksByStatus.Todo,
			&report.TasksByStatus.InProgress,
			&report.TasksByStatus.Done,
			&averageCloseTime,
			&report.CommentsCount,
			&userID,
			&name,
			&closedTasks,
		); err != nil {
			return stats.Report{}, fmt.Errorf("scan team statistics: %w", err)
		}

		if averageCloseTime.Valid {
			value := averageCloseTime.Float64
			report.AverageCloseTimeSeconds = &value
		}

		if userID.Valid {
			report.TopAssignees = append(
				report.TopAssignees,
				stats.TopAssignee{
					UserID:      userID.Int64,
					Name:        name.String,
					ClosedTasks: closedTasks.Int64,
				},
			)
		}
	}

	if err := rows.Err(); err != nil {
		return stats.Report{}, fmt.Errorf("iterate team statistics: %w", err)
	}

	return report, nil
}
