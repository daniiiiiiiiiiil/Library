package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ReaderService struct {
	readerRepo      repository.ReaderRepository
	userRepo        repository.UserRepository
	transactionRepo repository.TransactionRepository
	logger          *zap.Logger
}

func NewReaderService(
	readerRepo repository.ReaderRepository,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	logger *zap.Logger,
) *ReaderService {
	return &ReaderService{
		readerRepo:      readerRepo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

func (r *ReaderService) CreateReader(ctx context.Context, conn *pgx.Conn, reader *domain.Reader, password string) (*domain.Reader, error) {
	r.logger.Info("create reader started", zap.String("email", reader.Email), zap.String("name", reader.Name))

	if err := reader.ValidateReader(); err != nil {
		r.logger.Warn("reader validation failed", zap.String("email", reader.Email), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrValidation",
			Message: "Не прошли валидацию" + err.Error(),
		}
	}

	existsEmail, err := r.readerRepo.ExistsEmail(ctx, conn, reader.Email)
	if err != nil {
		r.logger.Error("failed to check email existence", zap.String("email", reader.Email), zap.Error(err))
		return nil, err
	}
	if existsEmail {
		r.logger.Warn("email already exists", zap.String("email", reader.Email))
		return nil, errors.BusinessError{
			Code:    "ErrReaderAlreadyExists",
			Message: "Читатель с таким email уже существует",
		}
	}

	existsPhone, err := r.readerRepo.ExistsPhone(ctx, conn, reader.Phone)
	if err != nil {
		r.logger.Error("failed to check phone existence", zap.String("phone", reader.Phone), zap.Error(err))
		return nil, err
	}
	if existsPhone {
		r.logger.Warn("phone already exists", zap.String("phone", reader.Phone))
		return nil, errors.BusinessError{
			Code:    "ErrReaderAlreadyExists",
			Message: "Читатель с таким номером уже существует",
		}
	}

	reader.Status = "active"
	createReader, err := r.readerRepo.CreateReader(ctx, conn, reader)
	if err != nil {
		r.logger.Error("failed to create reader", zap.String("email", reader.Email), zap.Error(err))
		return nil, err
	}

	user := domain.User{
		Email:        reader.Email,
		PasswordHash: password,
		Role:         "user",
		ReaderID:     &reader.Id,
	}

	if err := r.userRepo.CreateUser(ctx, conn, user); err != nil {
		r.logger.Error("failed to create user for reader", zap.Int("reader_id", reader.Id), zap.Error(err))
		return nil, err
	}

	r.logger.Info("reader created successfully", zap.Int("reader_id", createReader.Id), zap.String("email", createReader.Email))
	return createReader, nil
}

func (r *ReaderService) GetReader(ctx context.Context, conn *pgx.Conn, id int) (*domain.Reader, error) {
	r.logger.Debug("get reader started", zap.Int("reader_id", id))

	reader, err := r.readerRepo.GetByID(ctx, conn, id)
	if err != nil {
		r.logger.Warn("reader not found", zap.Int("reader_id", id), zap.Error(err))
		return nil, err
	}

	r.logger.Debug("get reader finished", zap.Int("reader_id", reader.Id))
	return reader, nil
}

func (r *ReaderService) GetByEmail(ctx context.Context, conn *pgx.Conn, email string) (*domain.Reader, error) {
	r.logger.Debug("get reader by email started", zap.String("email", email))

	reader, err := r.readerRepo.GetByEmail(ctx, conn, email)
	if err != nil {
		r.logger.Warn("reader not found by email", zap.String("email", email), zap.Error(err))
		return nil, err
	}

	r.logger.Debug("get reader by email finished", zap.Int("reader_id", reader.Id))
	return reader, nil
}

func (r *ReaderService) Update(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}) error {
	r.logger.Info("update reader started", zap.Int("reader_id", id))

	existingReader, err := r.readerRepo.GetByID(ctx, conn, id)
	if err != nil {
		r.logger.Warn("reader not found for update", zap.Int("reader_id", id), zap.Error(err))
		return errors.NotFoundError{
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
		r.logger.Warn("reader validation failed on update", zap.Int("reader_id", id), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrValidation",
			Message: "Не прошли валидацию" + err.Error(),
		}
	}

	if err := r.readerRepo.Update(ctx, conn, id, *existingReader); err != nil {
		r.logger.Error("failed to update reader", zap.Int("reader_id", id), zap.Error(err))
		return err
	}

	r.logger.Info("reader updated successfully", zap.Int("reader_id", id))
	return nil
}

func (r *ReaderService) Delete(ctx context.Context, conn *pgx.Conn, id int) error {
	r.logger.Info("delete reader started", zap.Int("reader_id", id))

	_, err := r.readerRepo.GetByID(ctx, conn, id)
	if err != nil {
		r.logger.Warn("reader not found for delete", zap.Int("reader_id", id), zap.Error(err))
		return errors.NotFoundError{
			Entity: "Reader",
			ID:     id,
		}
	}

	activeBooks, err := r.readerRepo.GetActiveBooksCount(ctx, conn, id)
	if err != nil {
		r.logger.Error("failed to get active books count", zap.Int("reader_id", id), zap.Error(err))
		return err
	}
	if activeBooks > 0 {
		r.logger.Warn("reader has active books, cannot delete", zap.Int("reader_id", id), zap.Int("active_books", activeBooks))
		return errors.BusinessError{
			Code:    "ErrReaderHasActiveBooks",
			Message: fmt.Sprintf("Нельзя удалить читателя, у него есть %d активных книг", activeBooks),
		}
	}

	hasDebt, err := r.readerRepo.HasDebt(ctx, conn, id)
	if err != nil {
		r.logger.Error("failed to check reader debt", zap.Int("reader_id", id), zap.Error(err))
		return err
	}
	if hasDebt {
		r.logger.Warn("reader has debt, cannot delete", zap.Int("reader_id", id))
		return errors.BusinessError{
			Code:    "ErrReaderHasDebtBook",
			Message: "Нельзя удалить читателя, у него есть долги",
		}
	}

	if err := r.userRepo.DeleteByReaderID(ctx, conn, id); err != nil {
		r.logger.Error("failed to delete user by reader id", zap.Int("reader_id", id), zap.Error(err))
		return err
	}

	if err := r.readerRepo.Delete(ctx, conn, id); err != nil {
		r.logger.Error("failed to delete reader", zap.Int("reader_id", id), zap.Error(err))
		return err
	}

	r.logger.Info("reader deleted successfully", zap.Int("reader_id", id))
	return nil
}

func (r *ReaderService) ListReader(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error) {
	limit, offset = limitOffset(limit, offset)
	r.logger.Debug("list readers started", zap.Int("limit", limit), zap.Int("offset", offset))

	reader, err := r.readerRepo.List(ctx, conn, limit, offset)
	if err != nil {
		r.logger.Error("failed to list readers", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrorGetListReader",
			Message: "Не удалось получить список читателей:" + err.Error(),
		}
	}

	r.logger.Debug("list readers finished", zap.Int("returned", len(reader)))
	return reader, nil
}

func (r *ReaderService) GetActiveReaders(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error) {
	limit, offset = limitOffset(limit, offset)
	r.logger.Debug("get active readers started", zap.Int("limit", limit), zap.Int("offset", offset))

	reader, err := r.readerRepo.GetActive(ctx, conn, limit, offset)
	if err != nil {
		r.logger.Error("failed to get active readers", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrorGetListReaderActive",
			Message: "Не удалось получить список активных читателей:" + err.Error(),
		}
	}

	r.logger.Debug("get active readers finished", zap.Int("returned", len(reader)))
	return reader, nil
}

func (r *ReaderService) GetDebtors(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error) {
	limit, offset = limitOffset(limit, offset)
	r.logger.Debug("get debtors started", zap.Int("limit", limit), zap.Int("offset", offset))

	reader, err := r.readerRepo.GetDebtors(ctx, conn, limit, offset)
	if err != nil {
		r.logger.Error("failed to get debtors", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrorGetListReaderDebtors",
			Message: "Не удалось получить список должников читателей:" + err.Error(),
		}
	}

	r.logger.Debug("get debtors finished", zap.Int("returned", len(reader)))
	return reader, nil
}

func (r *ReaderService) BlockReader(ctx context.Context, conn *pgx.Conn, readerId int) error {
	r.logger.Info("block reader started", zap.Int("reader_id", readerId))

	_, err := r.readerRepo.GetByID(ctx, conn, readerId)
	if err != nil {
		r.logger.Warn("reader not found for block", zap.Int("reader_id", readerId), zap.Error(err))
		return errors.NotFoundError{
			Entity: "Reader",
			ID:     readerId,
		}
	}

	if err := r.readerRepo.BlockReader(ctx, conn, readerId); err != nil {
		r.logger.Error("failed to block reader", zap.Int("reader_id", readerId), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrBlockReader",
			Message: "Не удалось заблокировать читателя:" + err.Error(),
		}
	}

	r.logger.Info("reader blocked successfully", zap.Int("reader_id", readerId))
	return nil
}

func (r *ReaderService) UnblockReader(ctx context.Context, conn *pgx.Conn, readerId int) error {
	r.logger.Info("unblock reader started", zap.Int("reader_id", readerId))

	_, err := r.readerRepo.GetByID(ctx, conn, readerId)
	if err != nil {
		r.logger.Warn("reader not found for unblock", zap.Int("reader_id", readerId), zap.Error(err))
		return errors.NotFoundError{
			Entity: "Reader",
			ID:     readerId,
		}
	}

	if err := r.readerRepo.UnBlockReader(ctx, conn, readerId); err != nil {
		r.logger.Error("failed to unblock reader", zap.Int("reader_id", readerId), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrUnBlockReader",
			Message: "Не удалось разблокировать читателя:" + err.Error(),
		}
	}

	r.logger.Info("reader unblocked successfully", zap.Int("reader_id", readerId))
	return nil
}

func (s *ReaderService) CanBorrow(ctx context.Context, conn *pgx.Conn, readerID int) error {
	s.logger.Debug("can borrow check started", zap.Int("reader_id", readerID))

	reader, err := s.readerRepo.GetByID(ctx, conn, readerID)
	if err != nil {
		s.logger.Warn("reader not found for can borrow", zap.Int("reader_id", readerID), zap.Error(err))
		return errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}

	if reader.Status == "blocked" {
		s.logger.Warn("reader is blocked", zap.Int("reader_id", readerID))
		return errors.BusinessError{
			Code:    "ErrReaderBlocked",
			Message: "Читатель заблокирован",
		}
	}

	if reader.BooksCount >= reader.MaxBooks {
		s.logger.Warn("reader reached max books limit", zap.Int("reader_id", readerID), zap.Int("books_count", reader.BooksCount), zap.Int("max_books", reader.MaxBooks))
		return errors.BusinessError{
			Code:    "ErrLimitExceeded",
			Message: fmt.Sprintf("Достигнут лимит книг (%d из %d)", reader.BooksCount, reader.MaxBooks),
		}
	}

	hasDebt, err := s.readerRepo.HasDebt(ctx, conn, readerID)
	if err != nil {
		s.logger.Error("failed to check reader debt", zap.Int("reader_id", readerID), zap.Error(err))
		return err
	}
	if hasDebt {
		s.logger.Warn("reader has debt", zap.Int("reader_id", readerID))
		return errors.BusinessError{
			Code:    "ErrReaderHasDebt",
			Message: "У читателя есть долги",
		}
	}

	s.logger.Debug("can borrow check passed", zap.Int("reader_id", readerID))
	return nil
}

func (r *ReaderService) GetReaderHistory(ctx context.Context, conn *pgx.Conn, readerId int, limit, offset int) ([]domain.Transaction, int, error) {
	limit, offset = limitOffset(limit, offset)
	r.logger.Debug("get reader history started", zap.Int("reader_id", readerId), zap.Int("limit", limit), zap.Int("offset", offset))

	reader, err := r.transactionRepo.ListByReader(ctx, conn, readerId, limit, offset)
	if err != nil {
		r.logger.Error("failed to get reader history", zap.Int("reader_id", readerId), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrGetReaderHistory",
			Message: "Не удалось получить историю читателя",
		}
	}

	_, total, err := r.transactionRepo.CountByReader(ctx, conn, readerId, limit, offset)
	if err != nil {
		r.logger.Error("failed to count reader history", zap.Int("reader_id", readerId), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "count_reader_history_error",
			Message: "Не удалось подсчитать историю читателя: " + err.Error(),
		}
	}

	r.logger.Debug("get reader history finished", zap.Int("reader_id", readerId), zap.Int("returned", len(reader)), zap.Int("total", total))
	return reader, total, nil
}
