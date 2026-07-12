package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/infrastructure/settings"
	"library/internal/repository"
	"library/pkg/errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type TransactionService struct {
	txRepo          repository.TransactionRepository
	copyRepo        repository.BookCopyRepository
	readerRepo      repository.ReaderRepository
	bookRepo        repository.BookRepository
	settingsRepo    repository.SettingRepository
	reservationRepo repository.ReservationRepository
}

func NewTransactionService(
	txRepo repository.TransactionRepository,
	copyRepo repository.BookCopyRepository,
	readerRepo repository.ReaderRepository,
	reservationRepo repository.ReservationRepository,
	settingRepo repository.SettingRepository,
) *TransactionService {
	return &TransactionService{
		txRepo:          txRepo,
		copyRepo:        copyRepo,
		readerRepo:      readerRepo,
		reservationRepo: reservationRepo,
		settingsRepo:    settingRepo,
	}
}

func (t *TransactionService) BorrowBook(ctx context.Context, conn *pgx.Conn, copyID, readerID int, dueDate time.Time) (*domain.Transaction, error) {
	// Проверяем, что копия существует и доступна
	copy, err := t.copyRepo.GetByID(ctx, conn, copyID)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "BookCopy",
			ID:     copyID,
		}
	}
	if copy.Status != "available" {
		return nil, errors.BusinessError{
			Code:    "copy_not_available",
			Message: fmt.Sprintf("Копия книги не доступна (статус: %s)", copy.Status),
		}
	}

	//  Проверяем, что читатель существует и активен
	reader, err := t.readerRepo.GetByID(ctx, conn, readerID)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}
	if reader.Status == "blocked" {
		return nil, errors.BusinessError{
			Code:    "reader_blocked",
			Message: "Читатель заблокирован",
		}
	}

	//  Проверяем, что читатель не превысил лимит книг
	if reader.BooksCount >= reader.MaxBooks {
		return nil, errors.BusinessError{
			Code:    "reader_limit_reached",
			Message: fmt.Sprintf("Достигнут лимит книг (%d из %d)", reader.BooksCount, reader.MaxBooks),
		}
	}

	//  Проверяем, что у читателя нет долгов
	hasDebt, err := t.readerRepo.HasDebt(ctx, conn, readerID)
	if err != nil {
		return nil, err
	}
	if hasDebt {
		return nil, errors.BusinessError{
			Code:    "reader_has_debt",
			Message: "У читателя есть долги",
		}
	}

	// Проверяем, что книга не зарезервирована другим читателем
	reserved, err := t.reservationRepo.IsBookReservedByOther(ctx, conn, copyID, readerID)
	if err != nil {
		return nil, err
	}
	if reserved {
		return nil, errors.BusinessError{
			Code:    "copy_reserved",
			Message: "Книга зарезервирована другим читателем",
		}
	}

	// Создаем транзакцию
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
		return nil, err
	}

	//Меняем статус копии на "borrowed"
	if err := t.copyRepo.UpdateStatus(ctx, conn, copyID, "borrowed"); err != nil {
		return nil, errors.BusinessError{
			Code:    "update_status_error",
			Message: "Не удалось обновить статус копии: " + err.Error(),
		}
	}

	//Увеличиваем счетчик книг у читателя
	if err := t.readerRepo.IncrementBookCount(ctx, conn, readerID); err != nil {
		return nil, errors.BusinessError{
			Code:    "increment_count_error",
			Message: "Не удалось обновить счетчик книг: " + err.Error(),
		}
	}

	return &transaction, nil
}

func (t *TransactionService) ReturnBook(ctx context.Context, conn *pgx.Conn, transactionID int, returnDate time.Time, fine float64) (*domain.Transaction, error) {
	tx, err := t.txRepo.GetByID(ctx, conn, transactionID)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Transaction",
			ID:     transactionID,
		}
	}
	isActive, err := t.txRepo.IsTransactionActive(ctx, conn, transactionID)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrIsTransactionActive",
			Message: "Не удалось проверить транзакция активна или нет" + err.Error(),
		}
	}

	if !isActive {
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
		return nil, errors.BusinessError{
			Code:    "ErrUpdateTransaction",
			Message: "Не удалось обновить транзакцию" + err.Error(),
		}
	}
	if err := t.copyRepo.UpdateStatus(ctx, conn, tx.CopyID, "available"); err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrUpdateStatus",
			Message: "Не удалось обновить статус у копии" + err.Error(),
		}
	}
	if err := t.readerRepo.DecrementBookCount(ctx, conn, tx.ReaderID); err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrDecrementBookCount",
			Message: "Не удалось вычесть 1 из счетчика книг у читателя" + err.Error(),
		}
	}
	return tx, nil
}

func (t *TransactionService) GetTransaction(ctx context.Context, conn *pgx.Conn, id int) (*domain.Transaction, error) {
	tx, err := t.txRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Transaction",
			ID:     id,
		}
	}
	return tx, nil
}

func (t *TransactionService) GetReaderTransactions(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Transaction, int, error) {
	limitOffset(limit, offset)
	tx, count, err := t.txRepo.CountByReader(ctx, conn, readerID, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrGetReaderTransactions",
			Message: "Не удалось получить список транзакций у читателя:" + err.Error(),
		}
	}
	return tx, count, nil
}

func (t *TransactionService) GetBookTransactions(ctx context.Context, conn *pgx.Conn, bookID, limit, offset int) ([]domain.Transaction, int, error) {
	limitOffset(limit, offset)
	tx, count, err := t.txRepo.CountByBook(ctx, conn, bookID, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrGetReaderTransactions",
			Message: "Не удалось получить список транзакций у читателя:" + err.Error(),
		}
	}
	return tx, count, nil
}

func (t *TransactionService) GetActiveReaderTransactions(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Transaction, error) {
	limitOffset(limit, offset)
	tx, err := t.txRepo.GetActiveByReader(ctx, conn, readerID, limit, offset)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}
	return tx, nil
}

func (t *TransactionService) GetOverdueTransactions(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Transaction, error) {
	limitOffset(limit, offset)
	tx, err := t.txRepo.GetOverdue(ctx, conn, limit, offset)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrGetOverdueTransactions",
			Message: "Не удалось получить все просроченные транзакции" + err.Error(),
		}
	}
	return tx, nil
}

func (t *TransactionService) ProcessOverdueTransactions(ctx context.Context, conn *pgx.Conn, limit, offset int) error {
	limitOffset(limit, offset)
	overdueTransactions, err := t.txRepo.GetOverdue(ctx, conn, limit, offset)
	if err != nil {
		return errors.BusinessError{
			Code:    "ErrGetOverdut",
			Message: "Не удалось получить просроченные транзакции: " + err.Error(),
		}
	}

	if len(overdueTransactions) == 0 {
		return nil
	}

	fineRate, err := t.settingsRepo.GetByKey(ctx, conn, "fine_rate_per_day")
	if err != nil {
		// Если настройки нет, используем значение по умолчанию
		fineRate = settings.Setting{Value: "1.50"}
	}

	rate, _ := strconv.ParseFloat(fineRate.Value, 64)

	//  Обрабатываем каждую просроченную транзакцию
	for _, tx := range overdueTransactions {
		fine := tx.CalculateFine(rate)

		tx.FineAmount = fine
		tx.Status = "overdue"

		if err := t.txRepo.Update(ctx, conn, tx); err != nil {
			return errors.BusinessError{
				Code:    "UpdateTransaction",
				Message: "Не удалось обновить транзакцию" + err.Error(),
			}
			// Логируем ошибку, но продолжаем с другими
			//continue
		}
		//пока что нету
		//// Отправляем уведомление читателю
		//if err := t.notificationService.NotifyOverdue(ctx, conn, tx.ReaderID, tx.ID, fine); err != nil {
		//
		//	continue
		//}

	}

	return nil
}

func (t *TransactionService) CalculateFine(ctx context.Context, conn *pgx.Conn, transactionID int) (float64, error) {
	tx, err := t.txRepo.GetByID(ctx, conn, transactionID)
	if err != nil {
		return 0, errors.NotFoundError{
			Entity: "Transaction",
			ID:     transactionID,
		}
	}

	if !tx.IsOverdue() {
		return 0, nil
	}

	fineRateSetting, err := t.settingsRepo.GetByKey(ctx, conn, "fine_rate_per_day")
	if err != nil {
		return tx.CalculateFine(1.50), nil
	}

	rate, err := strconv.ParseFloat(fineRateSetting.Value, 64)
	if err != nil {
		return 0, errors.BusinessError{
			Code:    "invalid_fine_rate",
			Message: "Неверный формат ставки штрафа в настройках",
		}
	}

	fine := tx.CalculateFine(rate)

	return fine, nil
}
