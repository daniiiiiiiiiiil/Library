package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"

	"github.com/jackc/pgx/v5"
)

type ReaderService struct {
	readerRepo      repository.ReaderRepository
	userRepo        repository.UserRepository
	transactionRepo repository.TransactionRepository
}

func NewReaderService(
	readerRepo repository.ReaderRepository,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
) *ReaderService {
	return &ReaderService{
		readerRepo:      readerRepo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
	}
}

func (r *ReaderService) CreateReader(ctx context.Context, conn *pgx.Conn, reader *domain.Reader, password string) (*domain.Reader, error) {
	if err := reader.ValidateReader(); err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrValidation",
			Message: "Не прошли валидацию" + err.Error(),
		}
	}

	existsEmail, err := r.readerRepo.ExistsEmail(ctx, conn, reader.Email)
	if err != nil {
		return nil, err
	}
	if existsEmail {
		return nil, errors.BusinessError{
			Code:    "ErrReaderAlreadyExists",
			Message: "Читатель с таким email уже существует",
		}
	}

	existsPhone, err := r.readerRepo.ExistsPhone(ctx, conn, reader.Phone)
	if err != nil {
		return nil, err
	}
	if existsPhone {
		return nil, errors.BusinessError{
			Code:    "ErrReaderAlreadyExists",
			Message: "Читатель с таким номером уже существует",
		}
	}
	reader.Status = "active"
	if err := r.readerRepo.Create(ctx, conn, *reader); err != nil {
		return nil, err
	}

	user := domain.User{
		Email:        reader.Email,
		PasswordHash: password,
		Role:         "user",
		ReaderID:     &reader.Id,
	}

	if err := r.userRepo.Create(ctx, conn, user); err != nil {
		return nil, err
	}

	return reader, nil
}

func (r *ReaderService) GetReader(ctx context.Context, conn *pgx.Conn, id int) (*domain.Reader, error) {
	reader, err := r.readerRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func (r *ReaderService) GetByEmail(ctx context.Context, conn *pgx.Conn, email string) (*domain.Reader, error) {
	reader, err := r.readerRepo.GetByEmail(ctx, conn, email)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func (r *ReaderService) Update(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}) (*domain.Reader, error) {
	existingReader, err := r.readerRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Reader",
			ID:     id,
		}
	}

	if name, ok := updates["name"].(string); ok {
		existingReader.Name = name
	}
	if phone, ok := updates["phone"].(string); ok {
		existingReader.Phone = phone
	}
	if email, ok := updates["email"].(string); ok {
		existingReader.Email = email
	}
	if status, ok := updates["status"].(string); ok {
		existingReader.Status = status
	}
	if maxBooks, ok := updates["max_books"].(int); ok {
		existingReader.MaxBooks = maxBooks
	}
	if booksCount, ok := updates["books_count"].(int); ok {
		existingReader.BooksCount = booksCount
	}

	if err := existingReader.ValidateReader(); err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrValidation",
			Message: "Не прошли валидацию" + err.Error(),
		}
	}

	if err := r.readerRepo.Update(ctx, conn, id, *existingReader); err != nil {
		return nil, err
	}
	return existingReader, nil
}

func (r *ReaderService) Delete(ctx context.Context, conn *pgx.Conn, id int) error {
	_, err := r.readerRepo.GetByID(ctx, conn, id)
	if err != nil {
		return errors.NotFoundError{
			Entity: "Reader",
			ID:     id,
		}
	}
	activeBooks, err := r.readerRepo.GetActiveBooksCount(ctx, conn, id)
	if err != nil {
		return err
	}
	if activeBooks > 0 {
		return errors.BusinessError{
			Code:    "ErrReaderHasActiveBooks",
			Message: fmt.Sprintf("Нельзя удалить читателя, у него есть %d активных книг", activeBooks),
		}
	}

	hasDebt, err := r.readerRepo.HasDebt(ctx, conn, id)
	if err != nil {
		return err
	}
	if hasDebt {
		return errors.BusinessError{
			Code:    "ErrReaderHasDebtBook",
			Message: "Нельзя удалить читателя, у него есть долги",
		}
	}

	if err := r.userRepo.DeleteByReaderID(ctx, conn, id); err != nil {
		return err
	}

	if err := r.readerRepo.Delete(ctx, conn, id); err != nil {
		return err
	}
	return nil
}

func (r *ReaderService) ListReader(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error) {
	limitOffset(limit, offset)
	reader, err := r.readerRepo.List(ctx, conn, limit, offset)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrorGetListReader",
			Message: "Не удалось получить список читателей:" + err.Error(),
		}
	}
	return reader, nil
}

func (r *ReaderService) GetActiveReaders(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error) {
	limitOffset(limit, offset)
	reader, err := r.readerRepo.GetActive(ctx, conn, limit, offset)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrorGetListReaderActive",
			Message: "Не удалось получить список активных читателей:" + err.Error(),
		}
	}
	return reader, nil
}

func (r *ReaderService) GetDebtors(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error) {
	limitOffset(limit, offset)
	reader, err := r.readerRepo.GetDebtors(ctx, conn, limit, offset)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrorGetListReaderDebtors",
			Message: "Не удалось получить список должников читателей:" + err.Error(),
		}
	}
	return reader, nil
}

func (r *ReaderService) BlockReader(ctx context.Context, conn *pgx.Conn, readerId int) error {
	_, err := r.readerRepo.GetByID(ctx, conn, readerId)
	if err != nil {
		return errors.NotFoundError{
			Entity: "Reader",
			ID:     readerId,
		}
	}
	if err := r.readerRepo.BlockReader(ctx, conn, readerId); err != nil {
		return errors.BusinessError{
			Code:    "ErrBlockReader",
			Message: "Не удалось заблокировать читателя:" + err.Error(),
		}
	}
	return nil
}

func (r *ReaderService) UnblockReader(ctx context.Context, conn *pgx.Conn, readerId int) error {
	_, err := r.readerRepo.GetByID(ctx, conn, readerId)
	if err != nil {
		return errors.NotFoundError{
			Entity: "Reader",
			ID:     readerId,
		}
	}
	if err := r.readerRepo.UnBlockReader(ctx, conn, readerId); err != nil {
		return errors.BusinessError{
			Code:    "ErrUnBlockReader",
			Message: "Не удалось разблокировать читателя:" + err.Error(),
		}
	}
	return nil
}

// internal/service/reader_service.go
func (s *ReaderService) CanBorrow(ctx context.Context, conn *pgx.Conn, readerID int) error {
	reader, err := s.readerRepo.GetByID(ctx, conn, readerID)
	if err != nil {
		return errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}

	if reader.Status == "blocked" {
		return errors.BusinessError{
			Code:    "ErrReaderBlocked",
			Message: "Читатель заблокирован",
		}
	}

	if reader.BooksCount >= reader.MaxBooks {
		return errors.BusinessError{
			Code:    "ErrLimitExceeded",
			Message: fmt.Sprintf("Достигнут лимит книг (%d из %d)", reader.BooksCount, reader.MaxBooks),
		}
	}

	hasDebt, err := s.readerRepo.HasDebt(ctx, conn, readerID)
	if err != nil {
		return err
	}
	if hasDebt {
		return errors.BusinessError{
			Code:    "ErrReaderHasDebt",
			Message: "У читателя есть долги",
		}
	}

	return nil
}

func (r *ReaderService) GetReaderHistory(ctx context.Context, conn *pgx.Conn, readerId int, limit, offset int) ([]domain.Transaction, int, error) {
	limitOffset(limit, offset)
	reader, err := r.transactionRepo.ListByReader(ctx, conn, readerId, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrGetReaderHistory",
			Message: "Не удалось получить историю читателя",
		}
	}
	_, total, err := r.transactionRepo.CountByReader(ctx, conn, readerId, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "count_reader_history_error",
			Message: "Не удалось подсчитать историю читателя: " + err.Error(),
		}
	}
	return reader, total, nil
}
