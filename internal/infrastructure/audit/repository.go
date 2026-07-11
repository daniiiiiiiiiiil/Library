package audit

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

func CreateAuditLog(ctx context.Context, conn *pgx.Conn, log AuditLog) error {
	sqlQuery := `
		INSERT INTO audit_log (user_id, action, entity_type, entity_id, log_timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := conn.Exec(ctx, sqlQuery,
		log.UserID,
		log.Action,
		log.EntityType,
		log.EntityID,
		time.Now(),
	)
	return err
}

func GetByIDAuditLog(ctx context.Context, conn *pgx.Conn, id int) (AuditLog, error) {
	sqlQuery := `
		SELECT audit_log_id, user_id, action, entity_type, entity_id, log_timestamp
		FROM audit_log
		WHERE audit_log_id = $1
	`
	var log AuditLog
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&log.ID,
		&log.UserID,
		&log.Action,
		&log.EntityType,
		&log.EntityID,
		&log.LogTimestamp,
	)
	if err != nil {
		return AuditLog{}, err
	}
	return log, nil
}

func ListAuditLogs(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]AuditLog, error) {
	sqlQuery := `
		SELECT audit_log_id, user_id, action, entity_type, entity_id, log_timestamp
		FROM audit_log
		ORDER BY log_timestamp DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.LogTimestamp,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func GetAuditLogsByEntity(ctx context.Context, conn *pgx.Conn, entityType string, entityID int, limit, offset int) ([]AuditLog, error) {
	sqlQuery := `
		SELECT audit_log_id, user_id, action, entity_type, entity_id, log_timestamp
		FROM audit_log
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY log_timestamp DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := conn.Query(ctx, sqlQuery, entityType, entityID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.LogTimestamp,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func GetAuditLogsByUser(ctx context.Context, conn *pgx.Conn, userID int, limit, offset int) ([]AuditLog, error) {
	sqlQuery := `
		SELECT audit_log_id, user_id, action, entity_type, entity_id, log_timestamp
		FROM audit_log
		WHERE user_id = $1
		ORDER BY log_timestamp DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := conn.Query(ctx, sqlQuery, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.LogTimestamp,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func GetAuditLogsByAction(ctx context.Context, conn *pgx.Conn, action string, limit, offset int) ([]AuditLog, error) {
	sqlQuery := `
		SELECT audit_log_id, user_id, action, entity_type, entity_id, log_timestamp
		FROM audit_log
		WHERE action = $1
		ORDER BY log_timestamp DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := conn.Query(ctx, sqlQuery, action, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.LogTimestamp,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}
