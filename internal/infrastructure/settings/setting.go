package settings

import (
	"library/pkg/errors"
	"time"
)

type Setting struct {
	ID          int        `json:"id"`
	Key         string     `json:"key"`
	Value       string     `json:"value"`
	Description string     `json:"description"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func NewSetting(
	id int,
	key string,
	value string,
	description string,
	updatedAt *time.Time,
) Setting {
	return Setting{
		ID:          id,
		Key:         key,
		Value:       value,
		Description: description,
		UpdatedAt:   updatedAt,
	}
}

func (s Setting) Validate() error {
	var errs errors.ValidationErrors

	if s.Key == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "key",
			Message: "ключ не может быть пустым",
		})
	} else if len(s.Key) > 100 {
		errs = append(errs, errors.ValidationError{
			Field:   "key",
			Message: "ключ не может превышать 100 символов",
		})
	}

	if s.Value == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "value",
			Message: "значение не может быть пустым",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}
