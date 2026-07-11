package domain

import (
	"fmt"
	"library/pkg/errors"
	"time"
)

type Reader struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email"`
	RegisteredAt time.Time `json:"registered_at"`
	Status       string    `json:"status"`
	BooksCount   int       `json:"books_count"`
	MaxBooks     int       `json:"max_books"`
}

func NewReader(
	id int,
	name string,
	phone string,
	email string,
	registeredAt time.Time,
	status string,
	bookCount int,
	maxBooks int) Reader {
	return Reader{
		Id:           id,
		Name:         name,
		Phone:        phone,
		Email:        email,
		RegisteredAt: registeredAt,
		Status:       status,
		BooksCount:   bookCount,
		MaxBooks:     maxBooks,
	}
}

func (r Reader) ValidateReader() error {
	var errs errors.ValidationErrors

	if r.Name == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "Name",
			Message: "Ваше имя не может быть пустым",
		})
	} else if len(r.Name) > 100 {
		errs = append(errs, errors.ValidationError{
			Field:   "Name",
			Message: fmt.Sprintf("Ваше имя не может превышать 100 символов (сейчас %d)", len(r.Name)),
		})
	}
	if r.Phone == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "Phone",
			Message: "Номер телефона не может быть пустым",
		})
	} else if len(r.Phone) > 30 {
		errs = append(errs, errors.ValidationError{
			Field:   "Phone",
			Message: fmt.Sprintf("Ваш номер телефона не может привышать 30 символов(сейчас %d)", len(r.Phone)),
		})
	}
	if len(r.Email) > 50 {
		errs = append(errs, errors.ValidationError{
			Field:   "Email",
			Message: fmt.Sprintf("Ваш email не может привысить 50 символов(сейчас %d)", len(r.Email)),
		})
	}
	if r.Status == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "Status",
			Message: "Ваш статус не может быть пустым",
		})
	} else if len(r.Status) > 20 {
		errs = append(errs, errors.ValidationError{
			Field:   "Status",
			Message: fmt.Sprintf("Ваш статус не может привышать 20 символов(сейчас %d)", len(r.Status)),
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// может ли взять книгу
func (r Reader) CanBorrow(id int) (bool, error) {
	if r.Id != id {
		return false, errors.ErrForbidden
	}

	if r.Status != "active" {
		return false, errors.NewBusinessError("reader_blocked", "читатель заблокирован")
	}

	return true, nil
}

func (r Reader) Block(id int) error {
	if r.Id != id {
		return errors.ErrForbidden
	}

	if r.Status == "blocked" {
		return errors.NewBusinessError("already_blocked", "читатель уже заблокирован")
	}

	r.Status = "blocked"
	return nil
}

func (r Reader) Unblock(id int) error {
	if r.Id != id {
		return errors.ErrForbidden
	}

	if r.Status != "blocked" {
		return errors.NewBusinessError("not_blocked", "читатель не заблокирован")
	}

	r.Status = "active"
	return nil
}
