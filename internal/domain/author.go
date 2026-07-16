package domain

import (
	"fmt"
	"library/pkg/errors"
	"time"
)

type Author struct {
	ID         int       `json:"id"`
	First_name string    `json:"first_name"`
	Last_name  string    `json:"last_name"`
	Biography  string    `json:"biography"`
	Birthday   time.Time `json:"birthday"`
}

func NewAuthor(id int, first_name string, last_name string, biography string, birthday time.Time) Author {
	return Author{
		ID:         id,
		First_name: first_name,
		Last_name:  last_name,
		Biography:  biography,
		Birthday:   birthday,
	}
}

func (a Author) ValidateAuthor() error {
	var errs errors.ValidationErrors
	if a.First_name == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "First_name",
			Message: "Имя автора не может быть пустым",
		})
	} else if len(a.First_name) > 100 {
		errs = append(errs, errors.ValidationError{
			Field:   "Name",
			Message: fmt.Sprintf("Имя автора не может быть больше 100 символов (сейчас %d)", len(a.First_name)),
		})
	}
	if a.Last_name == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "Last_name",
			Message: "Фамилия автора не может быть пустой",
		})
	} else if len(a.Last_name) > 100 {
		errs = append(errs, errors.ValidationError{
			Field:   "Last_name",
			Message: fmt.Sprintf("Фамилия автора не может быть больше 100 символов (сейчас %d)", len(a.Last_name)),
		})
	}
	if a.Biography == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "Biography",
			Message: "Биография автора не может быть пустой",
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}
