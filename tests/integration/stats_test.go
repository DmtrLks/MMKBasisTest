package stats_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"mmktestbasisByDGanichev/internal/config"
	"mmktestbasisByDGanichev/internal/database"
	mysqlrepository "mmktestbasisByDGanichev/internal/repository/mysql"
	"mmktestbasisByDGanichev/internal/task"

	"github.com/joho/godotenv"
)

func TestStatsRepositoryReport(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run integration tests")
	}

	_ = godotenv.Load("../../.env")

	db, err := database.NewMySQL(testDatabaseConfig())
	if err != nil {
		t.Fatalf("connect to MySQL: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := database.RunMigrations(ctx, database.NewClient(db)); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	now := time.Now().UTC().Truncate(time.Second)
	firstUserID := insertUser(t, ctx, tx, "Дмитрий")
	secondUserID := insertUser(t, ctx, tx, "Федор")
	targetTeamID := insertTeam(t, ctx, tx, firstUserID, "Бантики")
	foreignTeamID := insertTeam(t, ctx, tx, firstUserID, "Цветочки")

	insertTask(t, ctx, tx, targetTeamID, firstUserID, nil, task.TaskStatusTodo, now, nil)
	insertTask(t, ctx, tx, targetTeamID, firstUserID, nil, task.TaskStatusInProgress, now, nil)

	firstClosedAt := now.Add(-1 * time.Hour)
	firstDoneID := insertTask(
		t,
		ctx,
		tx,
		targetTeamID,
		firstUserID,
		&firstUserID,
		task.TaskStatusDone,
		firstClosedAt.Add(-1*time.Hour),
		&firstClosedAt,
	)

	secondClosedAt := now.Add(-2 * time.Hour)
	secondDoneID := insertTask(
		t,
		ctx,
		tx,
		targetTeamID,
		firstUserID,
		&firstUserID,
		task.TaskStatusDone,
		secondClosedAt.Add(-2*time.Hour),
		&secondClosedAt,
	)

	thirdClosedAt := now.Add(-3 * time.Hour)
	insertTask(
		t,
		ctx,
		tx,
		targetTeamID,
		firstUserID,
		&secondUserID,
		task.TaskStatusDone,
		thirdClosedAt.Add(-30*time.Minute),
		&thirdClosedAt,
	)

	insertComment(t, ctx, tx, firstDoneID, firstUserID, "Хорошо")
	insertComment(t, ctx, tx, firstDoneID, secondUserID, "Здорово")
	insertComment(t, ctx, tx, secondDoneID, firstUserID, "Великолепно")

	foreignClosedAt := now.Add(-1 * time.Hour)
	foreignTaskID := insertTask(
		t,
		ctx,
		tx,
		foreignTeamID,
		firstUserID,
		&firstUserID,
		task.TaskStatusDone,
		foreignClosedAt.Add(-10*time.Hour),
		&foreignClosedAt,
	)
	insertComment(t, ctx, tx, foreignTaskID, firstUserID, "Пу-пу-пуу....")

	repository := mysqlrepository.NewStatsRepository()
	report, err := repository.Report(ctx, tx, targetTeamID)
	if err != nil {
		t.Fatalf("get statistics report: %v", err)
	}

	if report.TeamID != targetTeamID {
		t.Fatalf("team ID: got %d, want %d", report.TeamID, targetTeamID)
	}

	if report.TasksByStatus.Todo != 1 ||
		report.TasksByStatus.InProgress != 1 ||
		report.TasksByStatus.Done != 3 {
		t.Fatalf("unexpected status counts: %+v", report.TasksByStatus)
	}

	if report.CommentsCount != 3 {
		t.Fatalf("comments count: got %d, want 3", report.CommentsCount)
	}

	if report.AverageCloseTimeSeconds == nil {
		t.Fatal("average close time is nil")
	}

	if *report.AverageCloseTimeSeconds != 4_200 {
		t.Fatalf("average close time: got %.0f, want 4200", *report.AverageCloseTimeSeconds)
	}

	if len(report.TopAssignees) != 2 {
		t.Fatalf("top assignees count: got %d, want 2", len(report.TopAssignees))
	}

	if report.TopAssignees[0].UserID != firstUserID ||
		report.TopAssignees[0].ClosedTasks != 2 {
		t.Fatalf("unexpected first assignee: %+v", report.TopAssignees[0])
	}

	if report.TopAssignees[1].UserID != secondUserID ||
		report.TopAssignees[1].ClosedTasks != 1 {
		t.Fatalf("unexpected second assignee: %+v", report.TopAssignees[1])
	}
}

func testDatabaseConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Driver:          env("DB_DRIVER", "mysql"),
		Host:            env("DB_HOST", "localhost"),
		Port:            env("DB_PORT", "3306"),
		Name:            env("DB_NAME", "task_manager"),
		User:            env("DB_USER", "app"),
		Password:        env("DB_PASSWORD", "app_password"),
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Minute,
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func insertUser(
	t *testing.T,
	ctx context.Context,
	tx *sql.Tx,
	name string,
) int64 {
	t.Helper()

	email := fmt.Sprintf("stats-%d-%s@example.com", time.Now().UnixNano(), name)

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)`,
		email,
		"integration-test",
		name,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	return lastInsertID(t, result)
}

func insertTeam(
	t *testing.T,
	ctx context.Context,
	tx *sql.Tx,
	createdBy int64,
	name string,
) int64 {
	t.Helper()

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO teams (name, created_by) VALUES (?, ?)`,
		name,
		createdBy,
	)
	if err != nil {
		t.Fatalf("insert team: %v", err)
	}

	return lastInsertID(t, result)
}

func insertTask(
	t *testing.T,
	ctx context.Context,
	tx *sql.Tx,
	teamID int64,
	createdBy int64,
	assigneeID *int64,
	status task.TaskStatus,
	createdAt time.Time,
	closedAt *time.Time,
) int64 {
	t.Helper()

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO tasks (
			team_id, title, description, status, created_by,
			assignee_id, created_at, updated_at, closed_at, version
		) VALUES (?, ?, '', ?, ?, ?, ?, ?, ?, 1)`,
		teamID,
		"Stats task",
		status,
		createdBy,
		assigneeID,
		createdAt,
		createdAt,
		closedAt,
	)
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	return lastInsertID(t, result)
}

func insertComment(
	t *testing.T,
	ctx context.Context,
	tx *sql.Tx,
	taskID int64,
	userID int64,
	content string,
) {
	t.Helper()

	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO task_comments (task_id, user_id, content) VALUES (?, ?, ?)`,
		taskID,
		userID,
		content,
	)
	if err != nil {
		t.Fatalf("insert comment: %v", err)
	}
}

func lastInsertID(t *testing.T, result sql.Result) int64 {
	t.Helper()

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("get last insert ID: %v", err)
	}

	return id
}
