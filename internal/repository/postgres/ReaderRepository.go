package postgres

import (
	"context"
	"library/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

type ReaderRepository struct{}

func (r *ReaderRepository) CreateReader(ctx context.Context, conn *pgx.Conn, reader *domain.Reader) (*domain.Reader, error) {
	sqlQuery := `
		INSERT INTO readers (name, phone, email, registered_at, status, max_books, books_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING reader_id
	`
	var id int
	err := conn.QueryRow(ctx, sqlQuery,
		reader.Name,
		reader.Phone,
		reader.Email,
		time.Now(),
		"active",
		reader.MaxBooks,
		0,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	reader.Id = id
	return reader, nil
}

func (r *ReaderRepository) GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.Reader, error) {
	sqlQuery := `
		SELECT reader_id, name, phone, email, registered_at, status, max_books, books_count
		FROM readers
		WHERE reader_id = $1
	`
	var reader domain.Reader
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&reader.Id,
		&reader.Name,
		&reader.Phone,
		&reader.Email,
		&reader.RegisteredAt,
		&reader.Status,
		&reader.MaxBooks,
		&reader.BooksCount,
	)
	if err != nil {
		return nil, err
	}
	return &reader, nil
}

func (r *ReaderRepository) GetByEmail(ctx context.Context, conn *pgx.Conn, email string) (*domain.Reader, error) {
	sqlQuery := `
		SELECT reader_id, name, phone, email, registered_at, status, max_books, books_count
		FROM readers
		WHERE email = $1
	`
	var reader domain.Reader
	err := conn.QueryRow(ctx, sqlQuery, email).Scan(
		&reader.Id,
		&reader.Name,
		&reader.Phone,
		&reader.Email,
		&reader.RegisteredAt,
		&reader.Status,
		&reader.MaxBooks,
		&reader.BooksCount,
	)
	if err != nil {
		return nil, err
	}
	return &reader, nil
}

func (r *ReaderRepository) Update(ctx context.Context, conn *pgx.Conn, id int, reader domain.Reader) (*domain.Reader, error) {
	sqlQuery := `
		UPDATE readers
		SET name = $1, phone = $2, email = $3, status = $4, max_books = $5, books_count = $6
		WHERE reader_id = $7
	`
	_, err := conn.Exec(ctx, sqlQuery,
		reader.Name,
		reader.Phone,
		reader.Email,
		reader.Status,
		reader.MaxBooks,
		reader.BooksCount,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &reader, nil
}

func (r *ReaderRepository) Delete(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
		DELETE FROM readers
		WHERE reader_id = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, id)
	return err
}

func (r *ReaderRepository) List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error) {
	sqlQuery := `
		SELECT reader_id, name, phone, email, registered_at, status, max_books, books_count
		FROM readers
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readers []domain.Reader
	for rows.Next() {
		var reader domain.Reader
		if err := rows.Scan(
			&reader.Id,
			&reader.Name,
			&reader.Phone,
			&reader.Email,
			&reader.RegisteredAt,
			&reader.Status,
			&reader.MaxBooks,
			&reader.BooksCount,
		); err != nil {
			return nil, err
		}
		readers = append(readers, reader)
	}
	return readers, nil
}

func (r *ReaderRepository) GetActive(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error) {
	sqlQuery := `
		SELECT reader_id, name, phone, email, registered_at, status, max_books, books_count
		FROM readers
		WHERE status = 'active'
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readers []domain.Reader
	for rows.Next() {
		var reader domain.Reader
		if err := rows.Scan(
			&reader.Id,
			&reader.Name,
			&reader.Phone,
			&reader.Email,
			&reader.RegisteredAt,
			&reader.Status,
			&reader.MaxBooks,
			&reader.BooksCount,
		); err != nil {
			return nil, err
		}
		readers = append(readers, reader)
	}
	return readers, nil
}

func (r *ReaderRepository) GetDebtors(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error) {
	sqlQuery := `
		SELECT DISTINCT
			r.reader_id,
			r.name,
			r.phone,
			r.email,
			r.registered_at,
			r.status,
			r.max_books,
			r.books_count
		FROM readers r
		JOIN transactions t ON r.reader_id = t.reader_id
		WHERE t.status = 'active' 
		  AND t.due_date < CURRENT_DATE
		ORDER BY r.name ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readers []domain.Reader
	for rows.Next() {
		var reader domain.Reader
		if err := rows.Scan(
			&reader.Id,
			&reader.Name,
			&reader.Phone,
			&reader.Email,
			&reader.RegisteredAt,
			&reader.Status,
			&reader.MaxBooks,
			&reader.BooksCount,
		); err != nil {
			return nil, err
		}
		readers = append(readers, reader)
	}
	return readers, nil
}

func (r *ReaderRepository) BlockReader(ctx context.Context, conn *pgx.Conn, readerID int) error {
	sqlQuery := `
		UPDATE readers
		SET status = 'blocked'
		WHERE reader_id = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, readerID)
	return err
}

func (r *ReaderRepository) UnBlockReader(ctx context.Context, conn *pgx.Conn, readerID int) error {
	sqlQuery := `
		UPDATE readers
		SET status = 'active'
		WHERE reader_id = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, readerID)
	return err
}

func (r *ReaderRepository) IncrementBookCount(ctx context.Context, conn *pgx.Conn, readerID int) error {
	sqlQuery := `
		UPDATE readers
		SET books_count = books_count + 1
		WHERE reader_id = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, readerID)
	return err
}

func (r *ReaderRepository) DecrementBookCount(ctx context.Context, conn *pgx.Conn, readerID int) error {
	sqlQuery := `
		UPDATE readers
		SET books_count = books_count - 1
		WHERE reader_id = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, readerID)
	return err
}

func (r *ReaderRepository) UpdateStatus(ctx context.Context, conn *pgx.Conn, readerID int, reader domain.Reader) error {
	sqlQuery := `
		UPDATE readers
		SET status = $1
		WHERE reader_id = $2
	`
	_, err := conn.Exec(ctx, sqlQuery, reader.Status, readerID)
	return err
}

func (r *ReaderRepository) Exists(ctx context.Context, conn *pgx.Conn, id int) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM readers WHERE reader_id = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *ReaderRepository) ExistsEmail(ctx context.Context, conn *pgx.Conn, email string) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM readers WHERE email = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *ReaderRepository) ExistsPhone(ctx context.Context, conn *pgx.Conn, phone string) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM readers WHERE phone = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, phone).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *ReaderRepository) GetActiveBooksCount(ctx context.Context, conn *pgx.Conn, readerID int) (int, error) {
	sqlQuery := `
		SELECT COUNT(*)
		FROM transactions
		WHERE reader_id = $1 AND status = 'active'
	`
	var count int
	err := conn.QueryRow(ctx, sqlQuery, readerID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ReaderRepository) HasDebt(ctx context.Context, conn *pgx.Conn, readerID int) (bool, error) {
	sqlQuery := `
		SELECT EXISTS (
			SELECT 1
			FROM transactions
			WHERE reader_id = $1 AND fine > 0 AND status = 'active'
		)
	`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, readerID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
