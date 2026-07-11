package domain

import (
	"fmt"
	"library/pkg/errors"
	"strings"
)

type Publisher struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

func NewPublisher(id int, name, address, phone string) Publisher {
	return Publisher{
		ID:      id,
		Name:    name,
		Address: address,
		Phone:   phone,
	}
}

func (p Publisher) Validate() error {
	var errs errors.ValidationErrors

	if strings.TrimSpace(p.Name) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "name",
			Message: "название издателя не может быть пустым",
		})
	} else if len(p.Name) > 100 {
		errs = append(errs, errors.ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("название издателя не может превышать 100 символов (сейчас %d)", len(p.Name)),
		})
	}

	if strings.TrimSpace(p.Address) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "address",
			Message: "адрес издателя не может быть пустым",
		})
	} else if len(p.Address) > 100 {
		errs = append(errs, errors.ValidationError{
			Field:   "address",
			Message: fmt.Sprintf("адрес издателя не может превышать 100 символов (сейчас %d)", len(p.Address)),
		})
	}

	if strings.TrimSpace(p.Phone) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "phone",
			Message: "телефон издателя не может быть пустым",
		})
	} else if len(p.Phone) > 30 {
		errs = append(errs, errors.ValidationError{
			Field:   "phone",
			Message: fmt.Sprintf("телефон издателя не может превышать 30 символов (сейчас %d)", len(p.Phone)),
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}
