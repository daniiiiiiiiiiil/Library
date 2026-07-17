package postgres

import (
	"context"
	"library/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

type SettingRepository struct{}

func (r *SettingRepository) CreateSetting(ctx context.Context, conn *pgx.Conn, setting domain.Setting) error {
	sqlQuery := `
		INSERT INTO settings (keysettings, valuesettings, description, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := conn.Exec(ctx, sqlQuery,
		setting.Key,
		setting.Value,
		setting.Description,
		time.Now(),
	)
	return err
}

func (r *SettingRepository) GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Setting, error) {
	sqlQuery := `
		SELECT setting_id, keysettings, valuesettings, description, updated_at
		FROM settings
		WHERE setting_id = $1
	`
	var setting domain.Setting
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&setting.ID,
		&setting.Key,
		&setting.Value,
		&setting.Description,
		&setting.UpdatedAt,
	)
	if err != nil {
		return domain.Setting{}, err
	}
	return setting, nil
}

func (r *SettingRepository) GetByKey(ctx context.Context, conn *pgx.Conn, key string) (domain.Setting, error) {
	sqlQuery := `
		SELECT setting_id, keysettings, valuesettings, description, updated_at
		FROM settings
		WHERE keysettings = $1
	`
	var setting domain.Setting
	err := conn.QueryRow(ctx, sqlQuery, key).Scan(
		&setting.ID,
		&setting.Key,
		&setting.Value,
		&setting.Description,
		&setting.UpdatedAt,
	)
	if err != nil {
		return domain.Setting{}, err
	}
	return setting, nil
}

func (r *SettingRepository) Update(ctx context.Context, conn *pgx.Conn, setting domain.Setting) error {
	sqlQuery := `
		UPDATE settings
		SET valuesettings = $1, description = $2, updated_at = $3
		WHERE setting_id = $4
	`
	_, err := conn.Exec(ctx, sqlQuery,
		setting.Value,
		setting.Description,
		time.Now(),
		setting.ID,
	)
	return err
}

func (r *SettingRepository) UpdateByKey(ctx context.Context, conn *pgx.Conn, key, value string) error {
	sqlQuery := `
		UPDATE settings
		SET valuesettings = $1, updated_at = $2
		WHERE keysettings = $3
	`
	_, err := conn.Exec(ctx, sqlQuery, value, time.Now(), key)
	return err
}

func (r *SettingRepository) Delete(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
		DELETE FROM settings
		WHERE setting_id = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, id)
	return err
}

func (r *SettingRepository) DeleteByKey(ctx context.Context, conn *pgx.Conn, key string) error {
	sqlQuery := `
		DELETE FROM settings
		WHERE keysettings = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, key)
	return err
}

func (r *SettingRepository) List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Setting, error) {
	sqlQuery := `
		SELECT setting_id, keysettings, valuesettings, description, updated_at
		FROM settings
		ORDER BY keysettings ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []domain.Setting
	for rows.Next() {
		var setting domain.Setting
		if err := rows.Scan(
			&setting.ID,
			&setting.Key,
			&setting.Value,
			&setting.Description,
			&setting.UpdatedAt,
		); err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}
	return settings, nil
}

func (r *SettingRepository) Exists(ctx context.Context, conn *pgx.Conn, key string) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM settings WHERE keysettings = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, key).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
