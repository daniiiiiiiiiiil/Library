package postgres

import (
	"context"
	"library/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

type TransactionRepository struct{}

func (r *TransactionRepository) CreateTransaction(ctx context.Context, conn *pgx.Conn, transaction domain.Transaction) error {
	sqlQuery := `
		INSERT INTO transactions(copy_id,reader_id,borrowed_at,due_date,returned_at,status,fine,types)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := conn.Exec(ctx, sqlQuery,
		transaction.CopyID,
		transaction.ReaderID,
		time.Now(),
		transaction.DueDate,
		transaction.ReturnedAt,
		transaction.Status,
		0,
		transaction.Types)
	return err
}

func (r *TransactionRepository) GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.Transaction, error) {
	sqlQuery := `
	SELECT transaction_id, copy_id, reader_id, borrowed_at, due_date, returned_at, status, fine, types
	FROM transactions 
	WHERE transaction_id = $1`
	var transaction domain.Transaction
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&transaction.ID,
		&transaction.CopyID,
		&transaction.ReaderID,
		&transaction.BorrowedAt,
		&transaction.DueDate,
		&transaction.ReturnedAt,
		&transaction.Status,
		&transaction.FineAmount,
		&transaction.Types)
	return &transaction, err
}

func (r *TransactionRepository) Update(ctx context.Context, conn *pgx.Conn, transaction domain.Transaction) error {
	sqlQuery := `
		UPDATE transactions
		SET copy_id = $1, reader_id = $2, due_date = $3, returned_at = $4, status = $5, fine = $6, types = $7
		WHERE transaction_id = $8`
	_, err := conn.Exec(ctx, sqlQuery,
		transaction.CopyID,
		transaction.ReaderID,
		transaction.DueDate,
		transaction.ReturnedAt,
		transaction.Status,
		transaction.FineAmount,
		transaction.Types,
		transaction.ID)
	return err
}

func (r *TransactionRepository) ListByReader(ctx context.Context, conn *pgx.Conn, readerID int, limit, offset int) ([]domain.Transaction, error) {
	sqlQuery := `
		SELECT transaction_id, copy_id, reader_id, borrowed_at, due_date, returned_at, status, fine, types
		FROM transactions
		WHERE reader_id = $1
		ORDER BY borrowed_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := conn.Query(ctx, sqlQuery, readerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.CopyID,
			&transaction.ReaderID,
			&transaction.BorrowedAt,
			&transaction.DueDate,
			&transaction.ReturnedAt,
			&transaction.Status,
			&transaction.FineAmount,
			&transaction.Types); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (r *TransactionRepository) ListByBook(ctx context.Context, conn *pgx.Conn, copyID int, limit, offset int) ([]domain.Transaction, error) {
	sqlQuery := `
		SELECT transaction_id, copy_id, reader_id, borrowed_at, due_date, returned_at, status, fine, types
		FROM transactions
		WHERE copy_id = $1
		ORDER BY borrowed_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := conn.Query(ctx, sqlQuery, copyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.CopyID,
			&transaction.ReaderID,
			&transaction.BorrowedAt,
			&transaction.DueDate,
			&transaction.ReturnedAt,
			&transaction.Status,
			&transaction.FineAmount,
			&transaction.Types); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (r *TransactionRepository) GetActiveByReader(ctx context.Context, conn *pgx.Conn, readerID int, limit, offset int) ([]domain.Transaction, error) {
	sqlQuery := `
		SELECT transaction_id, copy_id, reader_id, borrowed_at, due_date, returned_at, status, fine, types
		FROM transactions
		WHERE reader_id = $1 AND status = 'active'
		ORDER BY borrowed_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := conn.Query(ctx, sqlQuery, readerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.CopyID,
			&transaction.ReaderID,
			&transaction.BorrowedAt,
			&transaction.DueDate,
			&transaction.ReturnedAt,
			&transaction.Status,
			&transaction.FineAmount,
			&transaction.Types); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (r *TransactionRepository) GetOverdue(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Transaction, error) {
	sqlQuery := `
		SELECT transaction_id, copy_id, reader_id, borrowed_at, due_date, returned_at, status, fine, types
		FROM transactions
		WHERE due_date < NOW() AND status = 'active'
		ORDER BY due_date ASC
		LIMIT $1 OFFSET $2`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.CopyID,
			&transaction.ReaderID,
			&transaction.BorrowedAt,
			&transaction.DueDate,
			&transaction.ReturnedAt,
			&transaction.Status,
			&transaction.FineAmount,
			&transaction.Types); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (r *TransactionRepository) GetByCopyID(ctx context.Context, conn *pgx.Conn, copyID int, limit, offset int) ([]domain.Transaction, error) {
	sqlQuery := `
		SELECT transaction_id, copy_id, reader_id, borrowed_at, due_date, returned_at, status, fine, types
		FROM transactions
		WHERE copy_id = $1
		ORDER BY borrowed_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := conn.Query(ctx, sqlQuery, copyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.CopyID,
			&transaction.ReaderID,
			&transaction.BorrowedAt,
			&transaction.DueDate,
			&transaction.ReturnedAt,
			&transaction.Status,
			&transaction.FineAmount,
			&transaction.Types); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (r *TransactionRepository) ReturnBook(ctx context.Context, conn *pgx.Conn, transactionID int, returnDate time.Time, fine float64) error {
	sqlQuery := `
		UPDATE transactions
		SET status = 'completed', returned_at = $1, fine = $2
		WHERE transaction_id = $3`
	_, err := conn.Exec(ctx, sqlQuery, returnDate, fine, transactionID)
	return err
}

func (r *TransactionRepository) BorrowBook(ctx context.Context, conn *pgx.Conn, copyID, readerID int, dueDate time.Time) error {
	sqlQuery := `
        INSERT INTO transactions (copy_id, reader_id, borrowed_at, due_date, returned_at, status, fine, types)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	borrowedAt := time.Now()

	_, err := conn.Exec(ctx, sqlQuery,
		copyID,
		readerID,
		borrowedAt,
		dueDate,
		nil,
		"active",
		0,
		"borrow")
	return err
}

func (r *TransactionRepository) CountByReader(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Transaction, int, error) {
	sqlQuery := `
		SELECT transaction_id, copy_id, reader_id, borrowed_at, due_date, returned_at, status, fine, types
		FROM transactions
		WHERE reader_id = $1
		ORDER BY borrowed_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := conn.Query(ctx, sqlQuery, readerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.CopyID,
			&transaction.ReaderID,
			&transaction.BorrowedAt,
			&transaction.DueDate,
			&transaction.ReturnedAt,
			&transaction.Status,
			&transaction.FineAmount,
			&transaction.Types); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, transaction)
	}

	var count int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE reader_id = $1`, readerID).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return transactions, count, nil
}

func (r *TransactionRepository) CountByBook(ctx context.Context, conn *pgx.Conn, bookID, limit, offset int) ([]domain.Transaction, int, error) {
	sqlQuery := `
		SELECT transaction_id, copy_id, reader_id, borrowed_at, due_date, returned_at, status, fine, types
		FROM transactions
		WHERE copy_id = $1
		ORDER BY borrowed_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := conn.Query(ctx, sqlQuery, bookID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.CopyID,
			&transaction.ReaderID,
			&transaction.BorrowedAt,
			&transaction.DueDate,
			&transaction.ReturnedAt,
			&transaction.Status,
			&transaction.FineAmount,
			&transaction.Types); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, transaction)
	}

	var count int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE copy_id = $1`, bookID).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return transactions, count, nil
}

func (r *TransactionRepository) IsTransactionActive(ctx context.Context, conn *pgx.Conn, transactionID int) (bool, error) {
	sqlQuery := `
        SELECT EXISTS (
            SELECT 1 
            FROM transactions 
            WHERE transaction_id = $1 
              AND status = 'active'
        )
    `
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, transactionID).Scan(&exists)
	return exists, err
}

func (r *TransactionRepository) HasReaderBorrowedBook(ctx context.Context, conn *pgx.Conn, readerID, bookID int) (bool, error) {
	sqlQuery := `
		SELECT EXISTS (
			SELECT 1
			FROM transactions t
			JOIN book_copies bc ON t.copy_id = bc.book_copy_id
			WHERE t.reader_id = $1 AND bc.book_id = $2 AND t.status = 'completed'
		)
	`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, readerID, bookID).Scan(&exists)
	return exists, err
}
