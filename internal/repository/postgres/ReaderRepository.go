package postgres

import (
	"context"
	"fmt"
	"library/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

func CreateReader(ctx context.Context, conn pgx.Conn, r domain.Reader) error {
	sqlQuery := `
		INSERT INTO readers (name,phone,email,registered_at,status,max_books,books_count)
		VALUES ($1, $2, $3, $4, $5, $6,$7);
	`
	_, err := conn.Exec(ctx, sqlQuery, r.Name, r.Phone, r.Email, time.Now(), "active", r.MaxBooks, 0)

	return err
}

func GetByIDReader(ctx context.Context, conn pgx.Conn, bookID int) (*domain.Reader, error) {
	sqlQuery := `
		SELECT *
		FROM readers
		WHERE reader_id = $1
`
	var reader domain.Reader
	err := conn.QueryRow(ctx, sqlQuery, bookID).Scan(
		&reader.Id,
		&reader.Name,
		&reader.Phone,
		&reader.Email,
		&reader.RegisteredAt,
		&reader.Status,
		&reader.MaxBooks,
		&reader.BooksCount)
	return &reader, err
}

func GetByEmailReader(ctx context.Context, conn pgx.Conn, email string) (*domain.Reader, error) {
	sqlQuery := `
		SELECT *
		FROM readers
		WHERE email = $1`
	var reader domain.Reader
	err := conn.QueryRow(ctx, sqlQuery, email).Scan(
		&reader.Id,
		&reader.Name,
		&reader.Phone,
		&reader.Email,
		&reader.RegisteredAt,
		&reader.Status,
		&reader.MaxBooks,
		&reader.BooksCount)
	return &reader, err
}

func UpdateReader(ctx context.Context, conn pgx.Conn, bookID int, r domain.Reader) error {
	sqlQuery := `
			UPDATE readers
			SET name = $1,phone = $2,email = $3
			WHERE reader_id = $4`
	_, err := conn.Exec(ctx, sqlQuery, r.Name, r.Phone, r.Email, bookID)
	return err
}

func DeleteReader(ctx context.Context, conn pgx.Conn, bookID int) error {
	sqlQuery := `
		DELETE FROM readers
		WHERE reader_id = $1`
	_, err := conn.Exec(ctx, sqlQuery, bookID)
	return err
}

func ListReaders(ctx context.Context, conn pgx.Conn, limit, offset int) ([]domain.Reader, error) {
	sqlQuery := `
		SELECT *
		FROM readers
		ORDER BY name ASC
		LIMIT $1 OFFSET $2;
		`
	var readers []domain.Reader
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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
			&reader.BooksCount); err != nil {
			return nil, err
		}
		readers = append(readers, reader)
		printReader(reader)
	}
	return readers, nil
}

func printReader(readers domain.Reader) {
	fmt.Println("--------------------------------------------------")
	fmt.Println("ID:", readers.Id)
	fmt.Println("Name:", readers.Name)
	fmt.Println("Phone:", readers.Phone)
	fmt.Println("Email:", readers.Email)
	fmt.Println("RegisteredAt:", readers.RegisteredAt)
	fmt.Println("Status:", readers.Status)
	fmt.Println("Max Books:", readers.MaxBooks)
	fmt.Println("Books Count:", readers.BooksCount)
}

func GetActive(ctx context.Context, conn pgx.Conn, limit, offset int) (*[]domain.Reader, error) {
	sqlQuery := `
		SELECT *
		FROM readers
		WHERE status = 'active'
		ORDER BY name ASC
		LIMIT $1 OFFSET $2;`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return &[]domain.Reader{}, err
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
			&reader.BooksCount); err != nil {
			return &[]domain.Reader{}, err
		}
		readers = append(readers, reader)
		printReader(reader)
	}
	return &readers, nil
}

func GetDebtors(ctx context.Context, conn pgx.Conn, limit, offset int) (*[]domain.Reader, error) {
	sqlQuery := `
		SELECT DISTINCT
			r.reader_id,
			r.name,
			r.phone,
			r.email
		FROM readers r
		JOIN transactions t ON r.reader_id = t.reader_id
		WHERE t.status = 'active' 
		  AND t.due_date < CURRENT_DATE
		GROUP BY r.reader_id, r.name, r.phone, r.email
		LIMIT $1 OFFSET $2;`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return &[]domain.Reader{}, err
	}
	defer rows.Close()
	var readers []domain.Reader
	for rows.Next() {
		var reader domain.Reader
		if err := rows.Scan(&reader.Id,
			&reader.Name,
			&reader.Phone,
			&reader.Email,
			&reader.RegisteredAt,
			&reader.Status,
			&reader.MaxBooks,
			&reader.BooksCount); err != nil {
			return nil, err
		}
		readers = append(readers, reader)
		printReader(reader)
	}
	return &readers, nil
}

func BlockReader(ctx context.Context, conn pgx.Conn, readerId int) error {
	sqlQuery := `
	UPDATE readers
	SET status = 'blocked'
	WHERE reader_id = $1
`
	_, err := conn.Exec(ctx, sqlQuery, readerId)
	return err
}

func UnBlockReader(ctx context.Context, conn pgx.Conn, readerId int) error {
	sqlQuery := `
	UPDATE readers
	SET status = 'active'
	WHERE reader_id = $1
`
	_, err := conn.Exec(ctx, sqlQuery, readerId)
	return err
}

func IncrementBookCount(ctx context.Context, conn pgx.Conn, readerId int) error {
	sqlQuery := `
		UPDATE readers
		SET books_count = books_count + 1
		WHERE reader_id = $1`
	_, err := conn.Exec(ctx, sqlQuery, readerId)
	return err
}

func DecrementBookCount(ctx context.Context, conn pgx.Conn, readerId int) error {
	sqlQuery := `
		UPDATE readers
		SET books_count = books_count - 1
		WHERE reader_id = $1`
	_, err := conn.Exec(ctx, sqlQuery, readerId)
	return err
}

func UpdateStatusReader(ctx context.Context, conn pgx.Conn, readerID int, r domain.Reader) error {
	sqlQuery := `
		UPDATE readers
		SET status = $1
		WHERE reader_id = $2`
	_, err := conn.Exec(ctx, sqlQuery, r.Status, readerID)
	return err
}

func ExistsEmail(ctx context.Context, conn *pgx.Conn, email string) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT * FROM readers WHERE email = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
func ExistsPhone(ctx context.Context, conn *pgx.Conn, phone string) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM phone WHERE phone = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, phone).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GetActiveBooksCount(ctx context.Context, conn *pgx.Conn, readerID int) (int, error) {
	sqlQuery := `
        SELECT COUNT(*)
        FROM transactions
        WHERE reader_id = $1 AND status = 'active'
    `
	var count int
	err := conn.QueryRow(ctx, sqlQuery, readerID).Scan(&count)
	return count, err
}

func HasDebt(ctx context.Context, conn *pgx.Conn, readerID int) (bool, error) {
	sqlQuery := `
        SELECT EXISTS (
            SELECT 1
            FROM transactions
            WHERE reader_id = $1 AND fine > 0 AND status = 'active'
        )
    `
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, readerID).Scan(&exists)
	return exists, err
}

func Exists(ctx context.Context, conn *pgx.Conn, id int) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM readers WHERE reader_id = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(&exists)
	return exists, err
}
