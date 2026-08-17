package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mmktestbasisByDGanichev/internal/database"
	"mmktestbasisByDGanichev/internal/user"

	mysqldriver "github.com/go-sql-driver/mysql"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(
	ctx context.Context,
	db database.DBTX,
	usr user.User,
) (user.User, error) {
	const query = `
		INSERT INTO users (
			email,
			password_hash,
			name,
			created_at
		)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.ExecContext(ctx, query, usr.Email, usr.PasswordHash, usr.Name, usr.CreatedAt)
	if err != nil {
		if isDuplicateEntryError(err) {
			return user.User{}, user.ErrEmailAlreadyExists
		}

		return user.User{}, fmt.Errorf("insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return user.User{}, fmt.Errorf("get inserted user ID: %w", err)
	}

	usr.ID = id

	return usr, nil
}

func (r *UserRepository) FindByEmail(
	ctx context.Context,
	db database.DBTX,
	email string,
) (user.User, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			name,
			created_at
		FROM users
		WHERE email = ?
	`

	var u user.User

	err := db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Name,
		&u.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}

	if err != nil {
		return user.User{}, fmt.Errorf("find user by email: %w", err)
	}

	return u, nil
}

func isDuplicateEntryError(err error) bool {
	var mysqlErr *mysqldriver.MySQLError

	return errors.As(err, &mysqlErr) &&
		mysqlErr.Number == 1062
}

func (r *UserRepository) ExistsByID(
	ctx context.Context,
	db database.DBTX,
	userID int64,
) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = ?
		)
	`

	var exists bool

	if err := db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user existence: %w", err)
	}

	return exists, nil
}
