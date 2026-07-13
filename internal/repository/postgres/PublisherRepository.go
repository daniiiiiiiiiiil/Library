package postgres

import (
	"context"
	"fmt"
	"library/internal/domain"

	"github.com/jackc/pgx/v5"
)

func CreatePublisher(ctx context.Context, conn *pgx.Conn, publisher domain.Publisher) error {
	sqlQuery := `
		INSERT INTO publishers (name, address, phone)
		VALUES ($1, $2, $3)
	`
	_, err := conn.Exec(ctx, sqlQuery, publisher.Name, publisher.Address, publisher.Phone)
	return err
}

func GetByIDPublisher(ctx context.Context, conn *pgx.Conn, id int) (domain.Publisher, error) {
	sqlQuery := `
		SELECT publishers_id, name, address, phone
		FROM publishers
		WHERE publishers_id = $1
	`
	var publisher domain.Publisher
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&publisher.ID,
		&publisher.Name,
		&publisher.Address,
		&publisher.Phone,
	)
	if err != nil {
		return domain.Publisher{}, err
	}
	return publisher, nil
}

func UpdatePublisher(ctx context.Context, conn *pgx.Conn, id, publisher *domain.Publisher) error {
	sqlQuery := `
		UPDATE publishers
		SET name = $1, address = $2, phone = $3
		WHERE publishers_id = $4
	`
	_, err := conn.Exec(ctx, sqlQuery,
		publisher.Name,
		publisher.Address,
		publisher.Phone,
		id,
	)
	return err
}

func DeletePublisher(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
		DELETE FROM publishers
		WHERE publishers_id = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, id)
	return err
}

func ListPublishers(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Publisher, error) {
	sqlQuery := `
		SELECT publishers_id, name, address, phone
		FROM publishers
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var publishers []domain.Publisher
	for rows.Next() {
		var publisher domain.Publisher
		if err := rows.Scan(
			&publisher.ID,
			&publisher.Name,
			&publisher.Address,
			&publisher.Phone,
		); err != nil {
			return nil, err
		}
		publishers = append(publishers, publisher)
	}
	return publishers, nil
}

func ExistsPublisher(ctx context.Context, conn *pgx.Conn, id int) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM publishers WHERE publishers_id = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func SearchPublisher(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Publisher, int, error) {
	allowedColumns := map[string]bool{
		"name":    true,
		"address": true,
		"phone":   true,
	}
	if !allowedColumns[column] {
		return nil, 0, fmt.Errorf("недопустимая колонка: %s", column)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT publishers_id, name, address, phone
		FROM publishers
		WHERE %s ILIKE '%%' || $1 || '%%'
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`, column)

	rows, err := conn.Query(ctx, sqlQuery, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var publishers []domain.Publisher
	count := 0
	for rows.Next() {
		var publisher domain.Publisher
		if err := rows.Scan(
			&publisher.ID,
			&publisher.Name,
			&publisher.Address,
			&publisher.Phone,
		); err != nil {
			return nil, 0, err
		}
		publishers = append(publishers, publisher)
		count++
	}
	return publishers, count, nil
}
