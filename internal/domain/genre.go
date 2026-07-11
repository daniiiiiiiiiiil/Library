package domain

import (
	"fmt"
	"library/pkg/errors"
)

type Genre struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id"`
}

func NewGenre(id int, name string, parentID *int) Genre {
	return Genre{
		ID:       id,
		Name:     name,
		ParentID: parentID,
	}
}

func (g Genre) ValidateGenre() error {
	var errs errors.ValidationErrors
	if g.Name == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "Name",
			Message: "Имя не может быть пустым",
		})
	} else if len(g.Name) > 100 {
		errs = append(errs, errors.ValidationError{
			Field:   "Name",
			Message: fmt.Sprintf("Имя не может привышать 100 символов (сейчас %d)", len(g.Name)),
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}
