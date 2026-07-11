package audit

import (
	"library/pkg/errors"
	"time"
)

type AuditLog struct {
	ID           int       `json:"id"`
	UserID       *int      `json:"user_id,omitempty"`
	Action       string    `json:"action"`
	EntityType   string    `json:"entity_type"`
	EntityID     int       `json:"entity_id"`
	LogTimestamp time.Time `json:"log_timestamp"`
}

func NewAuditLog(
	id int,
	userID *int,
	action string,
	entityType string,
	entityID int,
	logTimestamp time.Time,
) AuditLog {
	return AuditLog{
		ID:           id,
		UserID:       userID,
		Action:       action,
		EntityType:   entityType,
		EntityID:     entityID,
		LogTimestamp: logTimestamp,
	}
}

func (a AuditLog) Validate() error {
	var errs errors.ValidationErrors

	if a.Action == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "action",
			Message: "действие не может быть пустым",
		})
	}

	if a.EntityType == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "entity_type",
			Message: "тип сущности не может быть пустым",
		})
	}

	if a.EntityID <= 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "entity_id",
			Message: "ID сущности должен быть положительным числом",
		})
	}

	if a.LogTimestamp.IsZero() {
		errs = append(errs, errors.ValidationError{
			Field:   "log_timestamp",
			Message: "время не может быть пустым",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}
