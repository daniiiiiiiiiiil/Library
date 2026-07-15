package dto

import (
	"library/internal/domain"
	"library/pkg/errors"
	"library/pkg/pagination"
	"strings"
	"time"
)

type CreateAuthorRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Biography string `json:"biography"`
	BirthDate string `json:"birth_date"`
}

func (r *CreateAuthorRequest) Validate() error {
	var errs errors.ValidationErrors

	if strings.TrimSpace(r.FirstName) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "first_name",
			Message: "Имя автора не может быть пустым",
		})
	}
	if strings.TrimSpace(r.LastName) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "last_name",
			Message: "Фамилия автора не может быть пустым",
		})
	}
	if strings.TrimSpace(r.Biography) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "biography",
			Message: "Биография автора не может быть пустым",
		})
	}
	layout := "2006-01-02"

	_, err := time.Parse(layout, r.BirthDate)
	if err != nil {
		errs = append(errs, errors.ValidationError{
			Field:   "birth_date",
			Message: "Неправильный формат даты, должен быть гггг-мм-дд(2006-01-02)" + err.Error(),
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *CreateAuthorRequest) ToDomain() domain.Author {
	return domain.NewAuthor(
		0,
		r.FirstName,
		r.LastName,
		r.Biography,
		r.BirthDate,
	)
}

func (r *CreateAuthorRequest) FromDomain(domain domain.Author) AuthorResponse {
	return AuthorResponse{
		id:         domain.ID,
		first_name: domain.First_name,
		last_name:  domain.Last_name,
		biography:  domain.Biography,
		birth_date: domain.Birthday,
	}
}

type UpdateAuthorRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Biography *string `json:"biography"`
	BirthDate *string `json:"birth_date"`
}

func (r *UpdateAuthorRequest) ToDomain(authorID int) (domain.Author, error) {
	if authorID < 0 {
		return domain.Author{}, errors.ValidationError{
			Field:   "ID",
			Message: "ID не может быть меньше 0",
		}
	}
	return domain.NewAuthor(
		authorID,
		"",
		"",
		"",
		"",
	), nil
}

type AuthorResponse struct {
	id         int    `json:"ID"`
	first_name string `json:"first_name"`
	last_name  string `json:"last_name"`
	biography  string `json:"biography"`
	birth_date string `json:"birth_date"`
}

type AuthorListResponse struct {
	authors    []AuthorResponse      `json:"authors"`
	pagination pagination.Pagination `json:"pagination"`
}

type SearchAuthorRequest struct {
	query  string `json:"query"`
	limit  int    `json:"limit"`
	offset int    `json:"offset"`
}
