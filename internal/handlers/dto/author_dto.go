package dto

import (
	"library/internal/domain"
	"library/pkg/errors"
	"library/pkg/pagination"
	"strings"
	"time"
)

type CreateAuthorRequest struct {
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Biography string     `json:"biography"`
	BirthDate *time.Time `json:"birth_date"`
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
		*r.BirthDate,
	)
}

func (r *CreateAuthorRequest) FromDomain(domain domain.Author) AuthorResponse {
	return AuthorResponse{
		ID:        domain.ID,
		FirstName: domain.First_name,
		LastName:  domain.Last_name,
		Biography: domain.Biography,
		BirthDate: domain.Birthday,
	}
}

type UpdateAuthorRequest struct {
	FirstName *string    `json:"first_name"`
	LastName  *string    `json:"last_name"`
	Biography *string    `json:"biography"`
	BirthDate *time.Time `json:"birth_date"`
}

func (r *UpdateAuthorRequest) Validate() error {
	var errs errors.ValidationErrors

	// Проверяем только если поле не nil
	if r.FirstName != nil && strings.TrimSpace(*r.FirstName) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "first_name",
			Message: "Имя не может быть пустым",
		})
	}
	if r.LastName != nil && strings.TrimSpace(*r.LastName) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "last_name",
			Message: "Фамилия не может быть пустым",
		})
	}
	if r.Biography != nil && strings.TrimSpace(*r.Biography) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "biography",
			Message: "Биография не может быть пустой",
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *UpdateAuthorRequest) ToDomain(authorID int, existingAuthor domain.Author) (domain.Author, error) {
	if authorID < 0 {
		return domain.Author{}, errors.ValidationError{
			Field:   "ID",
			Message: "ID не может быть меньше 0",
		}
	}

	birthday := existingAuthor.Birthday
	if r.BirthDate != nil {
		birthday = *r.BirthDate
	}

	firstName := existingAuthor.First_name
	if r.FirstName != nil {
		firstName = *r.FirstName
	}

	lastName := existingAuthor.Last_name
	if r.LastName != nil {
		lastName = *r.LastName
	}

	biography := existingAuthor.Biography
	if r.Biography != nil {
		biography = *r.Biography
	}

	return domain.NewAuthor(
		authorID,
		firstName,
		lastName,
		biography,
		birthday,
	), nil
}

type AuthorResponse struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Biography string    `json:"biography"`
	BirthDate time.Time `json:"birth_date"`
}

type AuthorListResponse struct {
	Authors    []AuthorResponse      `json:"authors"`
	Pagination pagination.Pagination `json:"pagination"`
}

type SearchAuthorRequest struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

func AuthorResponseFromDomain(author domain.Author) AuthorResponse {
	return AuthorResponse{
		ID:        author.ID,
		FirstName: author.First_name,
		LastName:  author.Last_name,
		Biography: author.Biography,
		BirthDate: author.Birthday,
	}
}

func NewAuthorListResponse(authors []domain.Author, total, limit, offset int) AuthorListResponse {
	resp := AuthorListResponse{
		Authors:    make([]AuthorResponse, 0, len(authors)),
		Pagination: pagination.NewPagination(total, limit, offset),
	}

	for _, author := range authors {
		resp.Authors = append(resp.Authors, AuthorResponseFromDomain(author))
	}

	return resp
}
