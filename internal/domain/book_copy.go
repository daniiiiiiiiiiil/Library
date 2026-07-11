package domain

import (
	"fmt"
	"library/pkg/errors"
	"time"
)

type BookCopy struct {
	ID         int        `json:"id"`
	BookID     int        `json:"book_id"`
	CopyNumber int        `json:"copy_number"`
	Status     string     `json:"status"`    // available, borrowed, reserved, damaged, lost
	Condition  string     `json:"condition"` // excellent, good, fair, poor, damaged
	ReaderID   *int       `json:"reader_id,omitempty"`
	BorrowedAt *time.Time `json:"borrowed_at,omitempty"`
}

func NewBookCopy(bookID int, copyNumber int, condition string) *BookCopy {
	return &BookCopy{
		BookID:     bookID,
		CopyNumber: copyNumber,
		Status:     "available",
		Condition:  condition,
		ReaderID:   nil,
		BorrowedAt: nil,
	}
}

func (bc *BookCopy) Validate() error {
	var errs errors.ValidationErrors

	if bc.BookID <= 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "book_id",
			Message: "ID книги должен быть положительным",
		})
	}

	if bc.CopyNumber <= 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "copy_number",
			Message: "номер копии должен быть положительным",
		})
	}

	if bc.Status == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "status",
			Message: "статус не может быть пустым",
		})
	}

	if bc.Condition == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "condition",
			Message: "состояние не может быть пустым",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// выдать книгу
func (bc *BookCopy) MarkAsBorrowed(readerID int) error {
	if bc.Status != "available" {
		return fmt.Errorf("книга не доступна (статус: %s)", bc.Status)
	}

	bc.Status = "borrowed"
	bc.ReaderID = &readerID
	now := time.Now()
	bc.BorrowedAt = &now

	return bc.Validate()
}

// сделать доступной
func (bc *BookCopy) MarkAsAvailable() error {
	bc.Status = "available"
	bc.ReaderID = nil
	bc.BorrowedAt = nil
	return bc.Validate()
}

// пометить как повреждённую
func (bc *BookCopy) MarkAsDamaged() error {
	bc.Status = "damaged"
	bc.ReaderID = nil
	bc.BorrowedAt = nil
	return bc.Validate()
}
