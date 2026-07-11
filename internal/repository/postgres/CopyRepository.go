package postgres

import (
	"context"
	"library/internal/domain"

	"github.com/jackc/pgx/v5"
)

func CreateCopy(ctx context.Context, conn *pgx.Conn, bookCopy domain.BookCopy) error {
	sqlQuery := `
INSERT INTO book_copies (book_id,copy_number,status,condition,reader_id,borrowed_at)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := conn.Exec(ctx, sqlQuery,
		bookCopy.BookID,
		bookCopy.CopyNumber,
		bookCopy.Status,
		bookCopy.Condition,
		bookCopy.ReaderID,
		bookCopy.BorrowedAt)
	return err
}

func GetByIDCopy(ctx context.Context, conn *pgx.Conn, id int) (domain.BookCopy, error) {
	sqlQuery := `
	SELECT *
	FROM book_copies
	WHERE book_copy_id = $1`

	var bookCopy domain.BookCopy
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&bookCopy.ID,
		&bookCopy.BookID,
		&bookCopy.CopyNumber,
		&bookCopy.Status,
		&bookCopy.Condition,
		&bookCopy.ReaderID,
		&bookCopy.BorrowedAt)
	if err != nil {
		return domain.BookCopy{}, err
	}
	return bookCopy, nil
}

func GetCopiesByBookID(ctx context.Context, conn *pgx.Conn, id int, limit, offset int) ([]domain.BookCopy, error) {
	sqlQuery := `
	SELECT *
	FROM book_copies
	WHERE book_id = $1
	LIMIT $2 OFFSET $3`
	rows, err := conn.Query(ctx, sqlQuery, id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bookCopies []domain.BookCopy
	for rows.Next() {
		var bookCopy domain.BookCopy
		if err := rows.Scan(
			&bookCopy.ID,
			&bookCopy.BookID,
			&bookCopy.CopyNumber,
			&bookCopy.Status,
			&bookCopy.Condition,
			&bookCopy.ReaderID,
			&bookCopy.BorrowedAt); err != nil {
			return nil, err
		}
		bookCopies = append(bookCopies, bookCopy)
	}
	return bookCopies, nil
}

func UpdateCopy(ctx context.Context, conn *pgx.Conn, bookCopy domain.BookCopy) error {
	sqlQuery := `
	UPDATE book_copies
	SET book_id = $1, copy_number = $2,status = $3,condition = $4,reader_id = $5,borrowed_at = $6
	WHERE book_copy_id = $7
`
	_, err := conn.Exec(ctx, sqlQuery,
		bookCopy.BookID,
		bookCopy.CopyNumber,
		bookCopy.Status,
		bookCopy.Condition,
		bookCopy.ReaderID,
		bookCopy.BorrowedAt,
		bookCopy.ID)
	return err
}

func DeleteCopy(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
		DELETE FROM book_copies
		WHERE book_copy_id = $1`
	_, err := conn.Exec(ctx, sqlQuery, id)
	return err
}

func GetAvailableCopy(ctx context.Context, conn *pgx.Conn, bookID int) (domain.BookCopy, error) {
	sqlQuery := `
		SELECT *
		FROM book_copies
		WHERE status = 'available' AND book_id = $1`

	var bookCopy domain.BookCopy
	if err := conn.QueryRow(ctx, sqlQuery, bookID).Scan(
		&bookCopy.ID,
		&bookCopy.BookID,
		&bookCopy.CopyNumber,
		&bookCopy.Status,
		&bookCopy.Condition,
		&bookCopy.ReaderID,
		&bookCopy.BorrowedAt); err != nil {
		return domain.BookCopy{}, err
	}
	return bookCopy, nil
}

func UpdateStatusCopy(ctx context.Context, conn *pgx.Conn, id int, status string) error {
	sqlQuery := `
	UPDATE book_copies
	SET status = $1
	WHERE book_copy_id = $2
`
	_, err := conn.Exec(ctx, sqlQuery, status, id)
	return err
}

func CountAvailable(ctx context.Context, conn *pgx.Conn, bookID int) (int, error) {
	sqlQuery := `
		SELECT COUNT(*)
		FROM book_copies
		WHERE book_id = $1 AND status = 'available'`

	var count int
	if err := conn.QueryRow(ctx, sqlQuery, bookID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func HasActiveCopies(ctx context.Context, conn *pgx.Conn, bookID int) (bool, error) {
	sqlQuery := `
        SELECT EXISTS (
            SELECT 1 
            FROM book_copies 
            WHERE book_id = $1 
              AND status IN ('borrowed', 'reserved')
        )
    `
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, bookID).Scan(&exists)
	return exists, err
}

func GetNextCopyNumber(ctx context.Context, conn *pgx.Conn, bookID int) (int, error) {
	sqlQuery := `
        SELECT COALESCE(MAX(copy_number), 0) + 1
        FROM book_copies
        WHERE book_id = $1
    `
	var nextNumber int
	err := conn.QueryRow(ctx, sqlQuery, bookID).Scan(&nextNumber)
	return nextNumber, err
}
