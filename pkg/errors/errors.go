package errors

import (
	"fmt"
	"strings"
)

var (
	ErrNotFound     = NewBusinessError("not_found", "запись не найдена")
	ErrConflict     = NewBusinessError("conflict", "конфликт данных")
	ErrUnauthorized = NewBusinessError("unauthorized", "не авторизован")
	ErrForbidden    = NewBusinessError("forbidden", "доступ запрещён")
	ErrInvalidInput = NewBusinessError("invalid_input", "некорректные входные данные")
	ErrInternal     = NewBusinessError("internal_error", "внутренняя ошибка сервера")
)

type BusinessError struct {
	Code    string
	Message string
}

func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

func (e BusinessError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e BusinessError) Is(target error) bool {
	if t, ok := target.(*BusinessError); ok {
		return e.Code == t.Code
	}
	return false
}

type NotFoundError struct {
	Entity string
	ID     int
}

func NewNotFoundError(entity string, id int) *NotFoundError {
	return &NotFoundError{
		Entity: entity,
		ID:     id,
	}
}

func NewNotFoundErrorf(entity string, id int, extra string) *NotFoundError {
	return &NotFoundError{
		Entity: entity,
		ID:     id,
	}
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s с ID %d не найден", e.Entity, e.ID)
}

type ConflictError struct {
	Entity  string
	Message string
}

func NewConflictError(entity, message string) *ConflictError {
	return &ConflictError{
		Entity:  entity,
		Message: message,
	}
}

func (e ConflictError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Entity, e.Message)
	}
	return fmt.Sprintf("%s уже существует", e.Entity)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	errs := make([]string, len(e))
	for i, err := range e {
		errs[i] = err.Error()
	}
	return strings.Join(errs, "; ")
}

func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

func (e *ValidationErrors) Add(field, message string) {
	*e = append(*e, ValidationError{
		Field:   field,
		Message: message,
	})
}

func (e ValidationErrors) ToMap() map[string]string {
	result := make(map[string]string)
	for _, err := range e {
		result[err.Field] = err.Message
	}
	return result
}

func (e ValidationErrors) ToMapWithPrefix(prefix string) map[string]string {
	result := make(map[string]string)
	for _, err := range e {
		key := err.Field
		if prefix != "" {
			key = prefix + "." + key
		}
		result[key] = err.Message
	}
	return result
}

func (e ValidationErrors) IsEmpty() bool {
	return len(e) == 0
}

func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

func IsConflictError(err error) bool {
	_, ok := err.(*ConflictError)
	return ok
}

func IsValidationError(err error) bool {
	_, ok := err.(ValidationErrors)
	return ok
}

func GetValidationErrors(err error) (ValidationErrors, bool) {
	if e, ok := err.(ValidationErrors); ok {
		return e, true
	}
	return nil, false
}
