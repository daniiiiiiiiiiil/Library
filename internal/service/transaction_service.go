package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type TransactionService struct {
	txRepo          repository.TransactionRepository
	copyRepo        repository.BookCopyRepository
	readerRepo      repository.ReaderRepository
	bookRepo        repository.BookRepository
	settingsRepo    repository.SettingRepository
	reservationRepo repository.ReservationRepository
	logger          *zap.Logger
}

func NewTransactionService(
	txRepo repository.TransactionRepository,
	copyRepo repository.BookCopyRepository,
	readerRepo repository.ReaderRepository,
	bookRepo repository.BookRepository,
	settingsRepo repository.SettingRepository,
	reservationRepo repository.ReservationRepository,
	logger *zap.Logger,
) *TransactionService {
	return &TransactionService{
		txRepo:          txRepo,
		copyRepo:        copyRepo,
		readerRepo:      readerRepo,
		bookRepo:        bookRepo,
		settingsRepo:    settingsRepo,
		reservationRepo: reservationRepo,
		logger:          logger,
	}
}

func (t *TransactionService) BorrowBook(ctx context.Context, conn *pgx.Conn, copyID, readerID int, dueDate time.Time) (*domain.Transaction, error) {
	t.logger.Info("borrow book started", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID), zap.String("due_date", dueDate.Format("2006-01-02")))

	copy, err := t.copyRepo.GetByID(ctx, conn, copyID)
	if err != nil {
		t.logger.Error("failed to get copy", zap.Int("copy_id", copyID), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "BookCopy",
			ID:     copyID,
		}
	}
	if copy.Status != "available" {
		t.logger.Warn("copy not available", zap.Int("copy_id", copyID), zap.String("status", copy.Status))
		return nil, errors.BusinessError{
			Code:    "copy_not_available",
			Message: fmt.Sprintf("Копия книги не доступна (статус: %s)", copy.Status),
		}
	}

	reader, err := t.readerRepo.GetByID(ctx, conn, readerID)
	if err != nil {
		t.logger.Error("failed to get reader", zap.Int("reader_id", readerID), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}
	if reader.Status == "blocked" {
		t.logger.Warn("reader is blocked", zap.Int("reader_id", readerID))
		return nil, errors.BusinessError{
			Code:    "reader_blocked",
			Message: "Читатель заблокирован",
		}
	}

	if reader.BooksCount >= reader.MaxBooks {
		t.logger.Warn("reader reached max books limit", zap.Int("reader_id", readerID), zap.Int("books_count", reader.BooksCount), zap.Int("max_books", reader.MaxBooks))
		return nil, errors.BusinessError{
			Code:    "reader_limit_reached",
			Message: fmt.Sprintf("Достигнут лимит книг (%d из %d)", reader.BooksCount, reader.MaxBooks),
		}
	}

	hasDebt, err := t.readerRepo.HasDebt(ctx, conn, readerID)
	if err != nil {
		t.logger.Error("failed to check reader debt", zap.Int("reader_id", readerID), zap.Error(err))
		return nil, err
	}
	if hasDebt {
		t.logger.Warn("reader has debt", zap.Int("reader_id", readerID))
		return nil, errors.BusinessError{
			Code:    "reader_has_debt",
			Message: "У читателя есть долги",
		}
	}

	reserved, err := t.reservationRepo.IsBookReservedByOther(ctx, conn, copyID, readerID)
	if err != nil {
		t.logger.Error("failed to check if book reserved by other", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID), zap.Error(err))
		return nil, err
	}
	if reserved {
		t.logger.Warn("book reserved by other reader", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID))
		return nil, errors.BusinessError{
			Code:    "copy_reserved",
			Message: "Книга зарезервирована другим читателем",
		}
	}

	transaction := domain.Transaction{
		CopyID:     copyID,
		ReaderID:   readerID,
		BorrowedAt: time.Now(),
		DueDate:    dueDate,
		ReturnedAt: nil,
		Status:     "active",
		FineAmount: 0,
		Types:      "borrow",
	}

	if err := t.txRepo.CreateTransaction(ctx, conn, transaction); err != nil {
		t.logger.Error("failed to create transaction", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID), zap.Error(err))
		return nil, err
	}

	_, err = t.copyRepo.UpdateStatus(ctx, conn, copyID, "borrowed")
	if err != nil {
		t.logger.Error("failed to update copy status", zap.Int("copy_id", copyID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "update_status_error",
			Message: "Не удалось обновить статус копии: " + err.Error(),
		}
	}

	if err := t.readerRepo.IncrementBookCount(ctx, conn, readerID); err != nil {
		t.logger.Error("failed to increment book count", zap.Int("reader_id", readerID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "increment_count_error",
			Message: "Не удалось обновить счетчик книг: " + err.Error(),
		}
	}

	t.logger.Info("book borrowed successfully", zap.Int("transaction_id", transaction.ID), zap.Int("copy_id", copyID), zap.Int("reader_id", readerID))
	return &transaction, nil
}

func (t *TransactionService) ReturnBook(ctx context.Context, conn *pgx.Conn, transactionID int, returnDate time.Time, fine float64) (*domain.Transaction, error) {
	t.logger.Info("return book started", zap.Int("transaction_id", transactionID), zap.String("return_date", returnDate.Format("2006-01-02")), zap.Float64("fine", fine))

	tx, err := t.txRepo.GetByID(ctx, conn, transactionID)
	if err != nil {
		t.logger.Error("failed to get transaction", zap.Int("transaction_id", transactionID), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Transaction",
			ID:     transactionID,
		}
	}

	isActive, err := t.txRepo.IsTransactionActive(ctx, conn, transactionID)
	if err != nil {
		t.logger.Error("failed to check if transaction active", zap.Int("transaction_id", transactionID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrIsTransactionActive",
			Message: "Не удалось проверить транзакция активна или нет" + err.Error(),
		}
	}

	if !isActive {
		t.logger.Warn("transaction not active", zap.Int("transaction_id", transactionID))
		return nil, errors.BusinessError{
			Code:    "transaction_not_active",
			Message: "Транзакция не активна",
		}
	}

	fineResult := tx.CalculateFine(fine)

	tx.ReturnedAt = &returnDate
	tx.FineAmount = fineResult
	if tx.IsOverdue() {
		tx.Status = "overdue"
	} else {
		tx.Status = "completed"
	}

	if err := t.txRepo.Update(ctx, conn, *tx); err != nil {
		t.logger.Error("failed to update transaction", zap.Int("transaction_id", transactionID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrUpdateTransaction",
			Message: "Не удалось обновить транзакцию" + err.Error(),
		}
	}

	_, err = t.copyRepo.UpdateStatus(ctx, conn, tx.CopyID, "available")
	if err != nil {
		t.logger.Error("failed to update copy status", zap.Int("copy_id", tx.CopyID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrUpdateStatus",
			Message: "Не удалось обновить статус у копии" + err.Error(),
		}
	}

	if err := t.readerRepo.DecrementBookCount(ctx, conn, tx.ReaderID); err != nil {
		t.logger.Error("failed to decrement book count", zap.Int("reader_id", tx.ReaderID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrDecrementBookCount",
			Message: "Не удалось вычесть 1 из счетчика книг у читателя" + err.Error(),
		}
	}

	t.logger.Info("book returned successfully", zap.Int("transaction_id", transactionID), zap.Float64("fine", tx.FineAmount), zap.String("status", tx.Status))
	return tx, nil
}

func (t *TransactionService) GetTransaction(ctx context.Context, conn *pgx.Conn, id int) (*domain.Transaction, error) {
	t.logger.Debug("get transaction started", zap.Int("transaction_id", id))

	tx, err := t.txRepo.GetByID(ctx, conn, id)
	if err != nil {
		t.logger.Warn("transaction not found", zap.Int("transaction_id", id), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Transaction",
			ID:     id,
		}
	}

	t.logger.Debug("get transaction finished", zap.Int("transaction_id", tx.ID))
	return tx, nil
}

func (t *TransactionService) GetReaderTransactions(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Transaction, int, error) {
	limit, offset = limitOffset(limit, offset)
	t.logger.Debug("get reader transactions started", zap.Int("reader_id", readerID), zap.Int("limit", limit), zap.Int("offset", offset))

	tx, count, err := t.txRepo.CountByReader(ctx, conn, readerID, limit, offset)
	if err != nil {
		t.logger.Error("failed to get reader transactions", zap.Int("reader_id", readerID), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrGetReaderTransactions",
			Message: "Не удалось получить список транзакций у читателя:" + err.Error(),
		}
	}

	t.logger.Debug("get reader transactions finished", zap.Int("reader_id", readerID), zap.Int("returned", len(tx)), zap.Int("total", count))
	return tx, count, nil
}

func (t *TransactionService) GetBookTransactions(ctx context.Context, conn *pgx.Conn, bookID, limit, offset int) ([]domain.Transaction, int, error) {
	limit, offset = limitOffset(limit, offset)
	t.logger.Debug("get book transactions started", zap.Int("book_id", bookID), zap.Int("limit", limit), zap.Int("offset", offset))

	tx, count, err := t.txRepo.CountByBook(ctx, conn, bookID, limit, offset)
	if err != nil {
		t.logger.Error("failed to get book transactions", zap.Int("book_id", bookID), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrGetReaderTransactions",
			Message: "Не удалось получить список транзакций у читателя:" + err.Error(),
		}
	}

	t.logger.Debug("get book transactions finished", zap.Int("book_id", bookID), zap.Int("returned", len(tx)), zap.Int("total", count))
	return tx, count, nil
}

func (t *TransactionService) GetActiveReaderTransactions(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Transaction, error) {
	limit, offset = limitOffset(limit, offset)
	t.logger.Debug("get active reader transactions started", zap.Int("reader_id", readerID), zap.Int("limit", limit), zap.Int("offset", offset))

	tx, err := t.txRepo.GetActiveByReader(ctx, conn, readerID, limit, offset)
	if err != nil {
		t.logger.Error("failed to get active reader transactions", zap.Int("reader_id", readerID), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}

	t.logger.Debug("get active reader transactions finished", zap.Int("reader_id", readerID), zap.Int("returned", len(tx)))
	return tx, nil
}

func (t *TransactionService) GetOverdueTransactions(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Transaction, error) {
	limit, offset = limitOffset(limit, offset)
	t.logger.Debug("get overdue transactions started", zap.Int("limit", limit), zap.Int("offset", offset))

	tx, err := t.txRepo.GetOverdue(ctx, conn, limit, offset)
	if err != nil {
		t.logger.Error("failed to get overdue transactions", zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrGetOverdueTransactions",
			Message: "Не удалось получить все просроченные транзакции" + err.Error(),
		}
	}

	t.logger.Debug("get overdue transactions finished", zap.Int("returned", len(tx)))
	return tx, nil
}

func (t *TransactionService) ProcessOverdueTransactions(ctx context.Context, conn *pgx.Conn, limit, offset int) error {
	limit, offset = limitOffset(limit, offset)
	t.logger.Info("process overdue transactions started", zap.Int("limit", limit), zap.Int("offset", offset))

	overdueTransactions, err := t.txRepo.GetOverdue(ctx, conn, limit, offset)
	if err != nil {
		t.logger.Error("failed to get overdue transactions", zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrGetOverdut",
			Message: "Не удалось получить просроченные транзакции: " + err.Error(),
		}
	}

	if len(overdueTransactions) == 0 {
		t.logger.Info("no overdue transactions found")
		return nil
	}

	fineRate, err := t.settingsRepo.GetByKey(ctx, conn, "fine_rate_per_day")
	if err != nil {
		t.logger.Warn("fine rate setting not found, using default", zap.Error(err))
		fineRate = domain.Setting{Value: "1.50"}
	}

	rate, _ := strconv.ParseFloat(fineRate.Value, 64)
	t.logger.Info("processing overdue transactions", zap.Int("count", len(overdueTransactions)), zap.Float64("fine_rate", rate))

	for _, tx := range overdueTransactions {
		fine := tx.CalculateFine(rate)
		tx.FineAmount = fine
		tx.Status = "overdue"

		if err := t.txRepo.Update(ctx, conn, tx); err != nil {
			t.logger.Error("failed to update overdue transaction", zap.Int("transaction_id", tx.ID), zap.Error(err))
			return errors.BusinessError{
				Code:    "UpdateTransaction",
				Message: "Не удалось обновить транзакцию" + err.Error(),
			}
		}

		//пока что нету
		//// Отправляем уведомление читателю
		//if err := t.notificationService.NotifyOverdue(ctx, conn, tx.ReaderID, tx.ID, fine); err != nil {
		//
		//	continue
		//}

		t.logger.Debug("overdue transaction processed", zap.Int("transaction_id", tx.ID), zap.Float64("fine", fine))
	}

	t.logger.Info("process overdue transactions completed", zap.Int("processed", len(overdueTransactions)))
	return nil
}

func (t *TransactionService) CalculateFine(ctx context.Context, conn *pgx.Conn, transactionID int) (float64, error) {
	t.logger.Debug("calculate fine started", zap.Int("transaction_id", transactionID))

	tx, err := t.txRepo.GetByID(ctx, conn, transactionID)
	if err != nil {
		t.logger.Error("failed to get transaction", zap.Int("transaction_id", transactionID), zap.Error(err))
		return 0, errors.NotFoundError{
			Entity: "Transaction",
			ID:     transactionID,
		}
	}

	if !tx.IsOverdue() {
		t.logger.Debug("transaction is not overdue", zap.Int("transaction_id", transactionID))
		return 0, nil
	}

	fineRateSetting, err := t.settingsRepo.GetByKey(ctx, conn, "fine_rate_per_day")
	if err != nil {
		t.logger.Warn("fine rate setting not found, using default", zap.Error(err))
		return tx.CalculateFine(1.50), nil
	}

	rate, err := strconv.ParseFloat(fineRateSetting.Value, 64)
	if err != nil {
		t.logger.Error("invalid fine rate format", zap.String("value", fineRateSetting.Value), zap.Error(err))
		return 0, errors.BusinessError{
			Code:    "invalid_fine_rate",
			Message: "Неверный формат ставки штрафа в настройках",
		}
	}

	fine := tx.CalculateFine(rate)
	t.logger.Debug("calculate fine finished", zap.Int("transaction_id", transactionID), zap.Float64("fine", fine))
	return fine, nil
}
