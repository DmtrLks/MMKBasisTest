package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mmktestbasisByDGanichev/internal/database"
	"mmktestbasisByDGanichev/internal/team"
)

type TeamRepository struct{}

func NewTeamRepository() *TeamRepository {
	return &TeamRepository{}
}

func (r *TeamRepository) Create(
	ctx context.Context,
	db database.DBTX,
	t team.Team,
) (team.Team, error) {
	const query = `
		INSERT INTO teams (
			name,
			created_by,
			created_at
		)
		VALUES (?, ?, ?)
	`

	result, err := db.ExecContext(ctx, query, t.Name, t.CreatedBy, t.CreatedAt)
	if err != nil {
		return team.Team{}, fmt.Errorf("insert team: %w", err)
	}

	teamID, err := result.LastInsertId()
	if err != nil {
		return team.Team{}, fmt.Errorf("get inserted team ID: %w", err)
	}

	t.ID = teamID

	return t, nil
}

func (r *TeamRepository) AddMember(
	ctx context.Context,
	db database.DBTX,
	member team.Member,
) error {
	const query = `
		INSERT INTO team_members (
			team_id,
			user_id,
			role
		)
		VALUES (?, ?, ?)
	`

	_, err := db.ExecContext(ctx, query, member.TeamID, member.UserID, member.Role)
	if err != nil {
		if isDuplicateEntryError(err) {
			return team.ErrAlreadyMember
		}

		return fmt.Errorf("insert team member: %w", err)
	}

	return nil
}

func (r *TeamRepository) ListByUserID(
	ctx context.Context,
	db database.DBTX,
	userID int64,
) ([]team.TeamWithRole, error) {
	const query = `
		SELECT
			t.id,
			t.name,
			t.created_by,
			t.created_at,
			tm.role
		FROM team_members AS tm
		INNER JOIN teams AS t
			ON t.id = tm.team_id
		WHERE tm.user_id = ?
		ORDER BY
			t.created_at DESC,
			t.id DESC
	`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list teams by user ID: %w", err)
	}
	defer rows.Close()

	result := make([]team.TeamWithRole, 0)

	for rows.Next() {
		var item team.TeamWithRole
		var role string

		if err := rows.Scan(
			&item.Team.ID,
			&item.Team.Name,
			&item.Team.CreatedBy,
			&item.Team.CreatedAt,
			&role,
		); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}

		item.Role = team.Role(role)

		if !item.Role.IsValid() {
			return nil, fmt.Errorf("invalid team role %q in database", role)
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate teams: %w", err)
	}

	return result, nil
}

func (r *TeamRepository) FindMember(
	ctx context.Context,
	db database.DBTX,
	teamID int64,
	userID int64,
) (team.Member, error) {
	return r.findMember(ctx, db, teamID, userID, false)
}

// FindMemberForUpdate блокирует строку участника до конца транзакции.
// Нужен там, где прочитанная роль тут же меняется, чтобы параллельные
// транзакции не затёрли изменения друг друга.
func (r *TeamRepository) FindMemberForUpdate(
	ctx context.Context,
	db database.DBTX,
	teamID int64,
	userID int64,
) (team.Member, error) {
	return r.findMember(ctx, db, teamID, userID, true)
}

func (r *TeamRepository) findMember(
	ctx context.Context,
	db database.DBTX,
	teamID int64,
	userID int64,
	forUpdate bool,
) (team.Member, error) {
	query := `
		SELECT role
		FROM team_members
		WHERE team_id = ?
		  AND user_id = ?
	`

	if forUpdate {
		query += " FOR UPDATE"
	}

	var role string

	err := db.QueryRowContext(ctx, query, teamID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return team.Member{}, team.ErrMemberNotFound
	}

	if err != nil {
		return team.Member{}, fmt.Errorf("find team member: %w", err)
	}

	member := team.Member{
		TeamID: teamID,
		UserID: userID,
		Role:   team.Role(role),
	}

	if !member.Role.IsValid() {
		return team.Member{}, fmt.Errorf("invalid team role %q in database", role)
	}

	return member, nil
}

func (r *TeamRepository) UpdateMemberRole(
	ctx context.Context,
	db database.DBTX,
	member team.Member,
) error {
	const query = `
		UPDATE team_members
		SET role = ?
		WHERE team_id = ?
		  AND user_id = ?
	`

	result, err := db.ExecContext(ctx, query, member.Role, member.TeamID, member.UserID)
	if err != nil {
		return fmt.Errorf("update team member role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get updated team member rows count: %w", err)
	}

	if rowsAffected == 0 {
		return team.ErrMemberNotFound
	}

	return nil
}

func (r *TeamRepository) IsMember(
	ctx context.Context,
	db database.DBTX,
	teamID int64,
	userID int64,
) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM team_members
			WHERE team_id = ?
			  AND user_id = ?
		)
	`

	var exists bool

	if err := db.QueryRowContext(
		ctx,
		query,
		teamID,
		userID,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check team membership: %w", err)
	}

	return exists, nil
}
