package dto

import (
	"library/internal/domain"
	"library/pkg/errors"
	"library/pkg/pagination"
	"strings"
	"time"
)

type CreateCopyRequest struct {
	Condition string `json:"condition"`
}

func (r *CreateCopyRequest) Validate() error {
	var errs errors.ValidationErrors

	if strings.TrimSpace(r.Condition) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "condition",
			Message: "condition не может быть пустым",
		})
	}

	// Проверка допустимых значений
	allowedConditions := []string{"excellent", "good", "fair", "poor", "damaged"}
	isValid := false
	for _, c := range allowedConditions {
		if r.Condition == c {
			isValid = true
			break
		}
	}
	if !isValid {
		errs = append(errs, errors.ValidationError{
			Field:   "condition",
			Message: "condition должно быть одним из: excellent, good, fair, poor, damaged",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *CreateCopyRequest) ToDomain(bookID, copyNumber int) domain.BookCopy {
	return *domain.NewBookCopy(
		bookID,
		"",
		copyNumber,
		r.Condition,
	)
}

type UpdateCopyRequest struct {
	Condition *string `json:"condition"`
	Status    *string `json:"status"`
}

func (r *UpdateCopyRequest) Validate() error {
	var errs errors.ValidationErrors
	if r.Condition != nil && strings.TrimSpace(*r.Condition) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "condition",
			Message: "condition не может быть пустым",
		})
	}
	if r.Status != nil && strings.TrimSpace(*r.Status) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "status",
			Message: "Статус не может быть пустым",
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

type UpdateCopyStatusRequest struct {
	Status   string `json:"status"`
	ReaderID *int   `json:"reader_id"`
}

func (r *UpdateCopyStatusRequest) Validate() error {
	var errs errors.ValidationErrors
	if strings.TrimSpace(r.Status) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "status",
			Message: "Статус не может быть пустым",
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

type CopyResponse struct {
	ID         int        `json:"ID"`
	BookID     int        `json:"book_id"`
	BookTitle  string     `json:"BookTitle"`
	CopyNumber int        `json:"copy_number"`
	Status     string     `json:"status"`
	Condition  string     `json:"condition"`
	ReaderId   *int       `json:"reader_id"`
	BorrowTime *time.Time `json:"borrow_time"`
}

type CopyListResponse struct {
	Copies     []CopyResponse        `json:"copies"`
	Pagination pagination.Pagination `json:"pagination"`
}

func CopyResponseFromDomain(copy domain.BookCopy) CopyResponse {
	return CopyResponse{
		ID:         copy.ID,
		BookID:     copy.BookID,
		CopyNumber: copy.CopyNumber,
		Status:     copy.Status,
		Condition:  copy.Condition,
		ReaderId:   copy.ReaderID,
	}
}
