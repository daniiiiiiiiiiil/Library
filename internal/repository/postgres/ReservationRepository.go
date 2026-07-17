package postgres

import (
	"context"
	"library/internal/domain"
	"library/internal/repository"

	"github.com/jackc/pgx/v5"
)

var _ repository.ReservationRepository = (*ReservationRepository)(nil)

type ReservationRepository struct{}

func (r *ReservationRepository) CreateReservation(ctx context.Context, conn *pgx.Conn, reserv *domain.Reservation) (*domain.Reservation, error) {
	sqlQuery := `
	INSERT INTO reservations (copy_id, reader_id, reserved_at, expires_at, status)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING reservation_id
	`
	var id int
	err := conn.QueryRow(ctx, sqlQuery,
		reserv.CopyID,
		reserv.ReaderID,
		reserv.ReservedAt,
		reserv.ExpiresAt,
		string(reserv.Status),
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	reserv.ID = id
	return reserv, nil
}

func (r *ReservationRepository) GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.Reservation, error) {
	sqlQuery := `
	SELECT reservation_id, copy_id, reader_id, reserved_at, expires_at, status
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
	reserve.Status = domain.ReservationStatus(reserve.Status)
	return &reserve, nil
}

func (r *ReservationRepository) GetActiveByReader(ctx context.Context, conn *pgx.Conn, readerID int, limit, offset int) ([]domain.Reservation, int, error) {
	sqlQuery := `
	SELECT reservation_id, copy_id, reader_id, reserved_at, expires_at, status
	FROM reservations
	WHERE reader_id = $1 AND status = 'active'
	LIMIT $2 OFFSET $3
	`
	rows, err := conn.Query(ctx, sqlQuery, readerID, limit, offset)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
		}
		reservations = append(reservations, reserve)
	}

	var total int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM reservations WHERE reader_id = $1 AND status = 'active'`, readerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return reservations, total, nil
}

func (r *ReservationRepository) GetActiveByCopy(ctx context.Context, conn *pgx.Conn, copyID int, limit, offset int) ([]domain.Reservation, int, error) {
	sqlQuery := `
	SELECT reservation_id, copy_id, reader_id, reserved_at, expires_at, status
	FROM reservations
	WHERE copy_id = $1 AND status = 'active'
	LIMIT $2 OFFSET $3
`
	rows, err := conn.Query(ctx, sqlQuery, copyID, limit, offset)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
		}
		reservations = append(reservations, reserve)
	}
	return reservations, len(reservations), nil
}

func (r *ReservationRepository) UpdateStatus(ctx context.Context, conn *pgx.Conn, id int, status string) error {
	sqlQuery := `
	UPDATE reservations
	SET status = $1
	WHERE reservation_id = $2
`
	_, err := conn.Exec(ctx, sqlQuery, status, id)
	return err
}

func (r *ReservationRepository) Delete(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
	DELETE FROM reservations
	WHERE reservation_id = $1
`
	_, err := conn.Exec(ctx, sqlQuery, id)
	return err
}

func (r *ReservationRepository) GetExpired(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reservation, error) {
	sqlQuery := `
	SELECT reservation_id, copy_id, reader_id, reserved_at, expires_at, status
	FROM reservations
	WHERE expires_at < NOW() AND status = 'active'
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

func (r *ReservationRepository) IsBookReservedByOther(ctx context.Context, conn *pgx.Conn, copyID, readerID int) (bool, error) {
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

func (r *ReservationRepository) HasActiveForCopy(ctx context.Context, conn *pgx.Conn, copyID int) (bool, error) {
	sqlQuery := `
		SELECT EXISTS (
			SELECT 1
			FROM reservations
			WHERE copy_id = $1 AND status = 'active'
		)
	`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, copyID).Scan(&exists)
	return exists, err
}
