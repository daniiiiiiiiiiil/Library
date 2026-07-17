package dto

import (
	"library/internal/domain"
	"library/pkg/errors"
	"library/pkg/pagination"
	"time"
)

type TransactionRequest struct {
	CopyID   int       `json:"CopyID"`
	ReaderID int       `json:"reader_id"`
	DueDate  time.Time `json:"due_date"`
}

func (r *TransactionRequest) Validate() error {
	var errs errors.ValidationErrors
	if r.CopyID < 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "CopyID",
			Message: "ID копии не может быть меньше 0",
		})
	}
	if r.ReaderID < 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "reader_id",
			Message: "ID читателя не может быть меньше 0",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *TransactionRequest) ToDomain() domain.Transaction {
	return domain.NewTransaction(
		0,
		r.CopyID,
		r.ReaderID,
		"borrow",
		time.Now(),
		r.DueDate,
		nil,
		"active",
		0,
	)
}

type ReturnBookRequest struct {
	TransactionID int       `json:"transaction_id"`
	ReturnDate    time.Time `json:"return_date"`
	FineAmount    float64
}

func (r *ReturnBookRequest) Validate() error {
	if r.TransactionID < 0 {
		return errors.ValidationError{
			Field:   "transaction_id",
			Message: "ID транзакции не может быть пустым",
		}
	}
	return nil
}

type TransactionResponse struct {
	ID          int        `json:"ID"`
	CopyID      int        `json:"CopyID"`
	BookTitle   string     `json:"BookTitle"`
	ReaderID    int        `json:"reader_id"`
	ReaderName  string     `json:"ReaderName"`
	Types       string     `json:"types"`
	BorrowedAt  time.Time  `json:"borrowed_at"`
	ReturnDate  *time.Time `json:"return_date"`
	Status      string     `json:"status"`
	FineAmount  float64    `json:"fine"`
	IsOverdue   bool       `json:"is_overdue"`
	DaysOverdue int        `json:"days_overdue"`
}

func TransactionRequestFromDomain(d domain.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          d.ID,
		CopyID:      d.CopyID,
		BookTitle:   "",
		ReaderID:    d.ReaderID,
		ReaderName:  "",
		Types:       d.Types,
		BorrowedAt:  d.BorrowedAt,
		ReturnDate:  d.ReturnedAt,
		Status:      d.Status,
		FineAmount:  d.FineAmount,
		IsOverdue:   d.IsOverdue(),
		DaysOverdue: d.CalculateDaysOverdue(),
	}
}

type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Pagination   pagination.Pagination `json:"pagination"`
}

func NewTransactionListResponse(
	transactions []domain.Transaction,
	bookTitles map[int]string,
	readerNames map[int]string,
	total, limit, offset int,
) TransactionListResponse {
	resp := TransactionListResponse{
		Transactions: make([]TransactionResponse, 0, len(transactions)),
		Pagination:   pagination.NewPagination(total, limit, offset),
	}

	for _, tx := range transactions {
		bookTitle := ""
		if title, ok := bookTitles[tx.CopyID]; ok {
			bookTitle = title
		}
		readerName := ""
		if name, ok := readerNames[tx.ReaderID]; ok {
			readerName = name
		}

		resp.Transactions = append(resp.Transactions, TransactionResponseFromDomain(
			tx,
			bookTitle,
			readerName,
		))
	}

	return resp
}

func TransactionResponseFromDomain(d domain.Transaction, bookTitle, readerName string) TransactionResponse {
	return TransactionResponse{
		ID:          d.ID,
		CopyID:      d.CopyID,
		BookTitle:   bookTitle,
		ReaderID:    d.ReaderID,
		ReaderName:  readerName,
		Types:       d.Types,
		BorrowedAt:  d.BorrowedAt,
		ReturnDate:  d.ReturnedAt,
		Status:      d.Status,
		FineAmount:  d.FineAmount,
		IsOverdue:   d.IsOverdue(),
		DaysOverdue: d.DaysOverdue(),
	}
}

type ProcessOverdueRequest struct {
	Limit int `json:"limit"`
}

func (r *ProcessOverdueRequest) Validate() error {
	if r.Limit < 0 {
		return errors.ValidationError{
			Field:   "limit",
			Message: "Лимит не может быть меньше 0",
		}
	}
	return nil
}
