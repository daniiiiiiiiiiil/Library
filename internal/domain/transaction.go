package domain

import (
	"library/pkg/errors"
	"time"
)

type Transaction struct {
	ID         int        `json:"id"`
	CopyID     int        `json:"copy_id"`
	ReaderID   int        `json:"reader_id"`
	Types      string     `json:"type"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	DueDate    time.Time  `json:"due_date"`
	ReturnedAt *time.Time `json:"returned_at"`
	Status     string     `json:"status"`
	FineAmount float64    `json:"fine_amount"`
}

func NewTransaction(
	id int,
	copyID int,
	readerID int,
	types string,
	borrowedAt time.Time,
	dueDate time.Time,
	returnedAt *time.Time,
	status string,
	fineAmount float64) Transaction {
	return Transaction{
		ID:         id,
		CopyID:     copyID,
		ReaderID:   readerID,
		Types:      types,
		BorrowedAt: borrowedAt,
		DueDate:    dueDate,
		ReturnedAt: returnedAt,
		Status:     status,
		FineAmount: fineAmount,
	}
}

func (t Transaction) ValidateTransaction() error {
	var errs errors.ValidationErrors

	if t.Types == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "types",
			Message: "тип транзакции не может быть пустым",
		})
	}

	if t.CopyID <= 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "copy_id",
			Message: "ID копии должен быть положительным числом",
		})
	}

	if t.ReaderID <= 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "reader_id",
			Message: "ID читателя должен быть положительным числом",
		})
	}

	if t.BorrowedAt.IsZero() {
		errs = append(errs, errors.ValidationError{
			Field:   "borrowed_at",
			Message: "дата выдачи не может быть пустой",
		})
	}

	if t.DueDate.IsZero() {
		errs = append(errs, errors.ValidationError{
			Field:   "due_date",
			Message: "дата возврата не может быть пустой",
		})
	}

	if !t.BorrowedAt.IsZero() && !t.DueDate.IsZero() {
		if t.DueDate.Before(t.BorrowedAt) {
			errs = append(errs, errors.ValidationError{
				Field:   "due_date",
				Message: "дата возврата не может быть раньше даты выдачи",
			})
		}
	}

	if t.Status == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "status",
			Message: "статус не может быть пустым",
		})
	}

	if t.FineAmount < 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "fine_amount",
			Message: "штраф не может быть отрицательным",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (t Transaction) CalculateFine(ratePerDay float64) float64 {
	day := t.BorrowedAt.Sub(t.DueDate)
	days := int(day.Hours() / 24)
	return float64(days) * ratePerDay
}

func (t *Transaction) IsOverdue() bool {
	if t.Status == "completed" {
		return false
	}
	if t.ReturnedAt != nil {
		return t.ReturnedAt.After(t.DueDate)
	}
	now := time.Now()
	return now.After(t.DueDate)
}

func (t *Transaction) Complete(returnedAt time.Time, ratePerDay float64) error {
	if t.Status == "completed" || t.Status == "overdue" {
		return errors.NewBusinessError("already_completed", "транзакция уже завершена")
	}

	if t.Status != "active" {
		return errors.NewBusinessError("invalid_status", "транзакция не активна")
	}

	if returnedAt.Before(t.BorrowedAt) {
		return errors.NewBusinessError("invalid_return_date",
			"дата возврата не может быть раньше даты выдачи")
	}

	now := time.Now().UTC()
	if returnedAt.After(now) {
		return errors.NewBusinessError("future_return_date",
			"дата возврата не может быть в будущем")
	}

	t.ReturnedAt = &returnedAt

	if returnedAt.After(t.DueDate) {
		t.Status = "overdue"
		t.FineAmount = t.CalculateFine(ratePerDay)
	} else {
		t.Status = "completed"
		t.FineAmount = 0
	}

	return nil
}

func (t Transaction) CalculateDaysOverdue() int {
	if t.Status == "completed" && !t.IsOverdue() {
		return 0
	}

	var compareDate time.Time
	if t.ReturnedAt != nil {
		compareDate = *t.ReturnedAt
	} else {
		compareDate = time.Now()
	}

	if compareDate.Before(t.DueDate) || compareDate.Equal(t.DueDate) {
		return 0
	}

	diff := compareDate.Sub(t.DueDate)
	days := int(diff.Hours() / 24)

	if diff.Hours() > float64(days*24) {
		days++
	}

	return days
}
