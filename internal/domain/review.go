package domain

import (
	"fmt"
	"library/pkg/errors"
	"strings"
	"time"
)

type Review struct {
	ID        int        `json:"id"`
	BookID    int        `json:"book_id"`
	ReaderID  int        `json:"reader_id"`
	Rating    float64    `json:"rating"`
	Comment   string     `json:"comment"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func NewReview(
	id int,
	bookID int,
	readerID int,
	rating float64,
	comment string,
	createdAt time.Time,
	updatedAt *time.Time,
) Review {
	return Review{
		ID:        id,
		BookID:    bookID,
		ReaderID:  readerID,
		Rating:    rating,
		Comment:   comment,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (r Review) Validate() error {
	var errs errors.ValidationErrors

	if r.BookID <= 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "book_id",
			Message: "ID книги должен быть положительным числом",
		})
	}

	if r.ReaderID <= 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "reader_id",
			Message: "ID читателя должен быть положительным числом",
		})
	}

	if r.Rating < 1 || r.Rating > 10 {
		errs = append(errs, errors.ValidationError{
			Field:   "rating",
			Message: fmt.Sprintf("рейтинг должен быть от 1 до 10 (текущий: %.1f)", r.Rating),
		})
	}

	if strings.TrimSpace(r.Comment) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "comment",
			Message: "комментарий не может быть пустым",
		})
	}

	if r.CreatedAt.IsZero() {
		errs = append(errs, errors.ValidationError{
			Field:   "created_at",
			Message: "дата создания не может быть пустой",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *Review) Update(rating float64, comment string) error {
	if rating >= 1 && rating <= 10 {
		r.Rating = rating
	}
	if strings.TrimSpace(comment) != "" {
		r.Comment = comment
	}
	now := time.Now()
	r.UpdatedAt = &now
	return r.Validate()
}
