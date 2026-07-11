package postgres

import (
	"context"
	"fmt"
	"library/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

func CreateTransaction(ctx context.Context, conn *pgx.Conn, transaction domain.Transaction) error {
	sqlQuery := `
		INSERT INTO transactions(copy_id,reader_id,borrowed_at,due_date,returned_at,status,fine,types)
		VALUES ($1, $2, $3, $4, $5, $6, $7,$8)`
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

func GetByIdTransaction(ctx context.Context, conn *pgx.Conn, id int) (*domain.Transaction, error) {
	sqlQuery := `
	SELECT * 
	FROM transactions 
	WHERE transaction_id=$1;`
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

func UpdateTransaction(ctx context.Context, conn *pgx.Conn, transaction domain.Transaction) error {
	sqlQuery := `
		UPDATE transactions
		SET copy_id = $1,reader_id = $2, due_date = $3,returned_at = $4,status = $5,fine = $6,types = $7
		WHERE transaction_id = $8;`
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

func ListByReader(ctx context.Context, conn *pgx.Conn, readerId int, limit, offset int) ([]domain.Transaction, error) {
	sqlQuery := `
		SELECT *
		FROM transactions
		WHERE reader_id=$1
		LIMIT $2 OFFSET $3;
		`
	var transactions []domain.Transaction
	rows, err := conn.Query(ctx, sqlQuery, readerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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
		printTransactions(transaction)
	}
	return transactions, nil
}

func printTransactions(transactions domain.Transaction) {
	fmt.Println("----------------------------")
	fmt.Println("ID", transactions.ID)
	fmt.Println("CopyID", transactions.CopyID)
	fmt.Println("ReaderID", transactions.ReaderID)
	fmt.Println("BorrowedAt", transactions.BorrowedAt)
	fmt.Println("DueDate", transactions.DueDate)
	fmt.Println("ReturnedAt", transactions.ReturnedAt)
	fmt.Println("Status", transactions.Status)
	fmt.Println("FineAmount", transactions.FineAmount)
	fmt.Println("Types", transactions.Types)
}

func ListByBook(ctx context.Context, conn *pgx.Conn, copyID int, limit, offset int) ([]domain.Transaction, error) {
	sqlQuery := `
		SELECT *
		FROM transactions
		WHERE copy_id=$1
		LIMIT $2 OFFSET $3;`
	var transactions []domain.Transaction
	rows, err := conn.Query(ctx, sqlQuery, copyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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
		printTransactions(transaction)
	}
	return transactions, nil
}

func GetActiveByReader(ctx context.Context, conn *pgx.Conn, readerId int, limit, offset int) ([]domain.Transaction, error) {
	sqlQuery := `
		SELECT *
		FROM transactions
		WHERE reader_id=$1 AND status='active'
		LIMIT $2 OFFSET $3;`
	var transactions []domain.Transaction
	rows, err := conn.Query(ctx, sqlQuery, readerId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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
		printTransactions(transaction)
	}
	return transactions, nil
}

func GetOverdue(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Transaction, error) {
	sqlQuery := `
		SELECT *
		FROM transactions
		WHERE due_date<NOW() AND status='active'
		LIMIT $1 OFFSET $2;`
	var transactions []domain.Transaction
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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
		printTransactions(transaction)
	}
	return transactions, nil
}

func GetByCopyID(ctx context.Context, conn *pgx.Conn, copyID int, limit, offset int) ([]domain.Transaction, error) {
	sqlQuery := `
			SELECT *
			FROM transactions
			WHERE copy_id=$1 AND status='active'
			LIMIT $2 OFFSET $3;`
	var transactions []domain.Transaction
	rows, err := conn.Query(ctx, sqlQuery, copyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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
		printTransactions(transaction)
	}
	return transactions, nil
}

func ReturnBook(ctx context.Context, conn *pgx.Conn, transactionId int, returnDate time.Time, fine float64) error {
	sqlQuery := `
		UPDATE transactions
		SET status = 'completed',returned_at=$1,fine = $2
		WHERE transaction_id=$3;`
	_, err := conn.Exec(ctx, sqlQuery, returnDate, fine, transactionId)
	return err
}

func BorrowBook(ctx context.Context, conn *pgx.Conn, copyID, readerID int, dueDate time.Time) error {
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

func CountByReader(ctx context.Context, conn *pgx.Conn, readerID int, limit, offset int) ([]domain.Transaction, int, error) {
	sqlQuery := `
        SELECT COUNT(*)
        FROM transactions
        WHERE reader_id = $1
        LIMIT $2 OFFSET $3
    `
	var count int
	err := conn.QueryRow(ctx, sqlQuery, readerID, limit, offset).Scan(&count)
	var transactions []domain.Transaction
	rows, err := conn.Query(ctx, sqlQuery, readerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
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
		printTransactions(transaction)
	}
	return transactions, count, err
}

func CountByBook(ctx context.Context, conn *pgx.Conn, bookID int, limit, offset int) ([]domain.Transaction, int, error) {
	sqlQuery := `
        SELECT COUNT(*)
        FROM transactions
        WHERE copy_id = $1
        LIMIT $2 OFFSET $3
    `
	var count int
	err := conn.QueryRow(ctx, sqlQuery, bookID, limit, offset).Scan(&count)
	var transactions []domain.Transaction
	rows, err := conn.Query(ctx, sqlQuery, bookID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
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
		printTransactions(transaction)
	}
	return transactions, count, err
}

func IsTransactionActive(ctx context.Context, conn *pgx.Conn, transactionID int) (bool, error) {
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
