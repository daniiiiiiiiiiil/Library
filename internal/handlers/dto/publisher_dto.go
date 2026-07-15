package dto

import (
	"library/internal/domain"
	"library/pkg/errors"
	"library/pkg/pagination"
	"strings"
)

type CreatePublisherRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

func (r *CreatePublisherRequest) Validate() error {
	var errs errors.ValidationErrors
	if strings.TrimSpace(r.Name) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "name",
			Message: "Имя не может быть пустым",
		})
	}
	if strings.TrimSpace(r.Address) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "address",
			Message: "Адрес не может быть пустым",
		})
	}
	if strings.TrimSpace(r.Phone) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "phone",
			Message: "Номер телефона не может быть пустым",
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *CreatePublisherRequest) ToDomain() domain.Publisher {
	return domain.NewPublisher(
		0,
		r.Name,
		r.Address,
		r.Phone,
	)
}

type UpdatePublisherRequest struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
	Phone   *string `json:"phone"`
}

func (r *UpdatePublisherRequest) Validate() error {
	var errs errors.ValidationErrors
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "name",
			Message: "Имя не может быть пустым",
		})
	}
	if r.Address != nil && strings.TrimSpace(*r.Address) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "address",
			Message: "Адрес не может быть пустым",
		})
	}
	if r.Phone != nil && strings.TrimSpace(*r.Phone) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "phone",
			Message: "Номер телефона не может быть пустым",
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *UpdatePublisherRequest) ToDomain(publisherID int) domain.Publisher {
	return domain.NewPublisher(
		publisherID,
		"",
		"",
		"",
	)
}

type PublisherResponse struct {
	ID         int    `json:"ID"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	Phone      string `json:"phone"`
	BooksCount int    `json:"books_count"`
}

type PublisherBooksResponse struct {
	publisher  PublisherResponse
	books      []BookResponse
	pagination pagination.Pagination
}
