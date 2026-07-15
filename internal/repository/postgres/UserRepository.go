package postgres

import (
	"context"
	"library/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

type UserRepository struct{}

func (r *UserRepository) CreateUser(ctx context.Context, conn *pgx.Conn, user domain.User) error {
	sqlQuery := `
		INSERT INTO users (email, password_hash, role, reader_id, created_at, updated_at, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := conn.Exec(ctx, sqlQuery,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.ReaderID,
		time.Now(),
		time.Now(),
		nil,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.User, error) {
	sqlQuery := `
		SELECT user_id, email, password_hash, role, reader_id, created_at, updated_at, last_login_at
		FROM users
		WHERE user_id = $1
	`
	var user domain.User
	var lastLoginAt *time.Time
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.ReaderID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err != nil {
		return domain.User{}, err
	}
	user.LastLoginAt = lastLoginAt
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, conn *pgx.Conn, email string) (domain.User, error) {
	sqlQuery := `
		SELECT user_id, email, password_hash, role, reader_id, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`
	var user domain.User
	var lastLoginAt *time.Time
	err := conn.QueryRow(ctx, sqlQuery, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.ReaderID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err != nil {
		return domain.User{}, err
	}
	user.LastLoginAt = lastLoginAt
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, conn *pgx.Conn, user domain.User) error {
	sqlQuery := `
		UPDATE users
		SET email = $1, role = $2, reader_id = $3, updated_at = $4
		WHERE user_id = $5
	`
	_, err := conn.Exec(ctx, sqlQuery,
		user.Email,
		user.Role,
		user.ReaderID,
		time.Now(),
		user.ID,
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
		DELETE FROM users
		WHERE user_id = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, id)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, conn *pgx.Conn, id int, newPasswordHash string) error {
	sqlQuery := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE user_id = $3
	`
	_, err := conn.Exec(ctx, sqlQuery, newPasswordHash, time.Now(), id)
	return err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
		UPDATE users
		SET last_login_at = $1
		WHERE user_id = $2
	`
	_, err := conn.Exec(ctx, sqlQuery, time.Now(), id)
	return err
}

func (r *UserRepository) DeleteByReaderID(ctx context.Context, conn *pgx.Conn, readerID int) error {
	sqlQuery := `
        DELETE FROM users
        WHERE reader_id = $1
    `
	_, err := conn.Exec(ctx, sqlQuery, readerID)
	return err
}
