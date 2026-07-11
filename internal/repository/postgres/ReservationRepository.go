package postgres

import (
	"context"
	"library/internal/domain"

	"github.com/jackc/pgx/v5"
)

func CreateReservation(ctx context.Context, conn *pgx.Conn, reserv domain.Reservation) error {
	sqlQuery := `
	INSERT INTO reservations (copy_id,reader_id,reserved_at,expires_at,status)
	VALUES ($1,$2,$3,$4,$5)
`
	_, err := conn.Exec(ctx, sqlQuery,
		reserv.CopyID,
		reserv.ReaderID,
		reserv.ReservedAt,
		reserv.ExpiresAt,
		reserv.Status)
	return err
}

func GetByIDReservation(ctx context.Context, conn *pgx.Conn, id int) (*domain.Reservation, error) {
	sqlQuery := `
	SELECT *
	FROM reservations
	WHERE reservation_id = $1
`
	var reserve domain.Reservation
	if err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&reserve.ID,
		&reserve.CopyID,
		&reserve.ReaderID,
		&reserve.ReservedAt,
		&reserve.ExpiresAt,
		&reserve.Status); err != nil {
		return nil, err
	}
	return &reserve, nil
}

func GetActiveByReaderReserv(ctx context.Context, conn *pgx.Conn, readerId int, limit, offset int) ([]domain.Reservation, error) {
	sqlQuery := `
	SELECT *
	FROM reservations
	WHERE reader_id = $1 AND status = 'active'
	LIMIT $2 OFFSET $3
`
	rows, err := conn.Query(ctx, sqlQuery, readerId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reservations []domain.Reservation
	for rows.Next() {
		var reserve domain.Reservation
		if err := rows.Scan(
			&reserve.ID,
			&reserve.CopyID,
			&reserve.ReaderID,
			&reserve.ReservedAt,
			&reserve.ExpiresAt,
			&reserve.Status); err != nil {
			return nil, err
		}
		reservations = append(reservations, reserve)
	}
	return reservations, nil
}

func GetActiveByBookReserv(ctx context.Context, conn *pgx.Conn, copyID int, limit, offset int) ([]domain.Reservation, error) {
	sqlQuery := `
	SELECT *
	FROM reservations
	WHERE copy_id = $1 AND status = 'active'
	LIMIT $2 OFFSET $3
`
	rows, err := conn.Query(ctx, sqlQuery, copyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reservations []domain.Reservation
	for rows.Next() {
		var reserve domain.Reservation
		if err := rows.Scan(
			&reserve.ID,
			&reserve.CopyID,
			&reserve.ReaderID,
			&reserve.ReservedAt,
			&reserve.ExpiresAt,
			&reserve.Status); err != nil {
			return nil, err
		}
		reservations = append(reservations, reserve)
	}
	return reservations, nil
}

func UpdateStatus(ctx context.Context, conn *pgx.Conn, id int, newStatus string) error {
	sqlQuery := `
	UPDATE reservations
	SET status = $1
	WHERE reservation_id = $2
`
	_, err := conn.Exec(ctx, sqlQuery, newStatus, id)
	return err
}

func DeleteReservation(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
	DELETE FROM reservations
	WHERE reservation_id = $1
`
	_, err := conn.Exec(ctx, sqlQuery, id)
	return err
}

func GetExpiredReserv(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reservation, error) {
	sqlQuery := `
	SELECT *
	FROM reservations
	WHERE expires_at < NOW() 
	LIMIT $1 OFFSET $2
`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reservations []domain.Reservation
	for rows.Next() {
		var reserve domain.Reservation
		if err := rows.Scan(
			&reserve.ID,
			&reserve.CopyID,
			&reserve.ReaderID,
			&reserve.ReservedAt,
			&reserve.ExpiresAt,
			&reserve.Status); err != nil {
			return nil, err
		}
		reservations = append(reservations, reserve)
	}
	return reservations, nil
}

func IsBookReservedByOther(ctx context.Context, conn *pgx.Conn, copyID, readerID int) (bool, error) {
	sqlQuery := `
		SELECT EXISTS (
			SELECT 1
			FROM reservations
			WHERE copy_id = $1 
			  AND status = 'active' 
			  AND reader_id != $2
		)
	`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, copyID, readerID).Scan(&exists)
	return exists, err
}
