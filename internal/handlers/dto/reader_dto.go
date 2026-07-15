package dto

import (
	"library/internal/domain"
	"library/pkg/errors"
	"library/pkg/pagination"
	"strings"
	"time"
)

type CreateReaderRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
	MaxBooks int    `json:"maxBooks"`
}

func (r *CreateReaderRequest) Validate() error {
	var errs errors.ValidationErrors
	if strings.TrimSpace(r.Name) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "name",
			Message: "Имя не может быть пустым",
		})
	}
	if strings.TrimSpace(r.Phone) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "phone",
			Message: "Номер телефона не может быть пустым",
		})
	}
	if strings.TrimSpace(r.Email) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "email",
			Message: "Email не может быть пустым",
		})
	}
	if len(r.Password) < 8 {
		errs = append(errs, errors.ValidationError{
			Field:   "password",
			Message: "Пароль не может быть меньше 8 символов",
		})
	}
	if r.MaxBooks < 0 || r.MaxBooks > 20 {
		errs = append(errs, errors.ValidationError{
			Field:   "maxBooks",
			Message: "Максимальное количество книг не может быть меньше 0 и больше 20",
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *CreateReaderRequest) ToDomain() domain.Reader {
	return domain.NewReader(
		0,
		r.Name,
		r.Phone,
		r.Email,
		time.Now(),
		"active",
		0,
		r.MaxBooks,
	)
}

type UpdateReaderRequest struct {
	Name     *string `json:"name"`
	Phone    *string `json:"phone"`
	Email    *string `json:"email"`
	MaxBooks *int    `json:"maxBooks"`
}

func (r *UpdateReaderRequest) ToDomain(readerID int, existingReader domain.Reader) (domain.Reader, error) {
	name := existingReader.Name
	if r.Name != nil {
		name = *r.Name
	}
	phone := existingReader.Phone
	if r.Phone != nil {
		phone = *r.Phone
	}
	email := existingReader.Email
	if r.Email != nil {
		email = *r.Email
	}
	maxBooks := existingReader.MaxBooks
	if r.MaxBooks != nil {
		maxBooks = *r.MaxBooks
	}

	return domain.NewReader(
		readerID,
		name,
		phone,
		email,
		existingReader.RegisteredAt,
		existingReader.Status,
		existingReader.BooksCount,
		maxBooks,
	), nil
}

type ReaderResponse struct {
	ID           int       `json:"ID"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email"`
	RegisteredAt time.Time `json:"registeredAt"`
	Status       string    `json:"status"`
	BooksCount   int       `json:"books_count"`
	MaxBooks     int       `json:"max_books"`
	ActiveBooks  int       `json:"active_books"`
}

func ReaderFromDomain(reader domain.Reader) ReaderResponse {
	return ReaderResponse{
		ID:           reader.Id,
		Name:         reader.Name,
		Phone:        reader.Phone,
		Email:        reader.Email,
		RegisteredAt: reader.RegisteredAt,
		Status:       reader.Status,
		BooksCount:   reader.BooksCount,
		MaxBooks:     reader.MaxBooks,
	}
}

type ReaderHistoryResponse struct {
	reader      ReaderResponse
	transaction []TransactionResponse
	pagination  pagination.Pagination
}
