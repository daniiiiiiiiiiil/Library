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

type CopyService struct {
	copyRepo        repository.BookCopyRepository
	bookRepo        repository.BookRepository
	readerRepo      repository.ReaderRepository
	reservationRepo repository.ReservationRepository
	logger          *zap.Logger
}

func NewCopyService(
	copyRepo repository.BookCopyRepository,
	bookRepo repository.BookRepository,
	readerRepo repository.ReaderRepository,
	reservationRepo repository.ReservationRepository,
	logger *zap.Logger,
) *CopyService {
	return &CopyService{
		copyRepo:        copyRepo,
		bookRepo:        bookRepo,
		readerRepo:      readerRepo,
		reservationRepo: reservationRepo,
		logger:          logger,
	}
}

func (c *CopyService) CreateCopy(ctx context.Context, conn *pgx.Conn, copy domain.BookCopy) error {
	c.logger.Info("create copy started", zap.Int("book_id", copy.BookID), zap.Int("copy_number", copy.CopyNumber), zap.String("condition", copy.Condition))

	if err := copy.Validate(); err != nil {
		c.logger.Warn("copy validation failed", zap.Int("book_id", copy.BookID), zap.Error(err))
		return errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации" + err.Error(),
		}
	}

	if copy.BookID > 0 {
		exists, err := c.bookRepo.Exists(ctx, conn, copy.BookID)
		if err != nil {
			c.logger.Error("failed to check book existence", zap.Int("book_id", copy.BookID), zap.Error(err))
			return errors.BusinessError{
				Code:    "ErrBookExists",
				Message: "Не удалось проверить существование книги" + err.Error(),
			}
		}
		if !exists {
			c.logger.Warn("book not found for copy creation", zap.Int("book_id", copy.BookID))
			return errors.NotFoundError{
				Entity: "BookNotFound",
				ID:     copy.BookID,
			}
		}
	}
	if err := c.copyRepo.CreateCopy(ctx, conn, &copy); err != nil {
		c.logger.Error("failed to create copy", zap.Int("book_id", copy.BookID), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrCreateCopy",
			Message: "Не удалось создать копию" + err.Error(),
		}
	}

	c.logger.Info("copy created successfully", zap.Int("copy_id", copy.ID), zap.Int("book_id", copy.BookID), zap.Int("copy_number", copy.CopyNumber))
	return nil
}

func (c *CopyService) GetCopy(ctx context.Context, conn *pgx.Conn, id int) (*domain.BookCopy, error) {
	c.logger.Debug("get copy started", zap.Int("copy_id", id))

	copyId, err := c.copyRepo.GetByID(ctx, conn, id)
	if err != nil {
		c.logger.Warn("copy not found", zap.Int("copy_id", id), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrGetCopy",
			Message: fmt.Sprintf("Не удалось получить копию %d по id", id) + err.Error(),
		}
	}

	c.logger.Debug("get copy finished", zap.Int("copy_id", copyId.ID))
	return copyId, nil
}

func (c *CopyService) GetCopiesByBook(ctx context.Context, conn *pgx.Conn, bookID, limit, offset int) ([]domain.BookCopy, error) {
	limit, offset = limitOffset(limit, offset)
	c.logger.Debug("get copies by book started", zap.Int("book_id", bookID), zap.Int("limit", limit), zap.Int("offset", offset))

	copys, err := c.copyRepo.GetCopiesByBookID(ctx, conn, bookID, limit, offset)
	if err != nil {
		c.logger.Error("failed to get copies by book", zap.Int("book_id", bookID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrGetCopy",
			Message: "Не удалось получить все копии книг" + err.Error(),
		}
	}

	c.logger.Debug("get copies by book finished", zap.Int("book_id", bookID), zap.Int("copies_count", len(copys)))
	return copys, nil
}

func (c *CopyService) UpdateCopy(ctx context.Context, conn *pgx.Conn, copy *domain.BookCopy) error {
	c.logger.Info("update copy started", zap.Int("copy_id", copy.ID), zap.Int("book_id", copy.BookID))

	if copy.ID > 0 {
		exists, err := c.copyRepo.ExistsCopy(ctx, conn, copy.ID)
		if err != nil {
			c.logger.Error("failed to check copy existence", zap.Int("copy_id", copy.ID), zap.Error(err))
			return errors.NotFoundError{
				Entity: "Copy",
				ID:     copy.ID,
			}
		}
		if !exists {
			c.logger.Warn("copy not found for update", zap.Int("copy_id", copy.ID))
			return errors.NotFoundError{
				Entity: "Copy",
				ID:     copy.ID,
			}
		}
	}
	if err := copy.Validate(); err != nil {
		c.logger.Warn("copy validation failed on update", zap.Int("copy_id", copy.ID), zap.Error(err))
		return errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации" + err.Error(),
		}
	}
	if copy.Status == "borrowed" {
		c.logger.Warn("cannot update borrowed copy", zap.Int("copy_id", copy.ID))
		return errors.BusinessError{
			Code:    "ErrCopyBorrowedCannotUpdate",
			Message: "Копия выдана,нельзя обновить выданную копию",
		}
	}
	if err := c.copyRepo.Update(ctx, conn, *copy); err != nil {
		c.logger.Error("failed to update copy", zap.Int("copy_id", copy.ID), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrUpdateCopy",
			Message: "Не удалось обновить копию" + err.Error(),
		}
	}

	c.logger.Info("copy updated successfully", zap.Int("copy_id", copy.ID))
	return nil
}

func (c *CopyService) DeleteCopy(ctx context.Context, conn *pgx.Conn, id int) (*domain.BookCopy, error) {
	c.logger.Info("delete copy started", zap.Int("copy_id", id))

	copyDel, err := c.copyRepo.GetByID(ctx, conn, id)
	if err != nil {
		c.logger.Warn("copy not found for delete", zap.Int("copy_id", id), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Copy",
			ID:     id,
		}
	}
	if copyDel.Status == "borrowed" {
		c.logger.Warn("cannot delete borrowed copy", zap.Int("copy_id", id))
		return nil, errors.BusinessError{
			Code:    "ErrCopyBorrowedCannotUpdate",
			Message: "Копия выдана,нельзя обновить выданную копию",
		}
	}

	hasReservations, err := c.reservationRepo.HasActiveForCopy(ctx, conn, id)
	if err != nil {
		c.logger.Error("failed to check active reservations for copy", zap.Int("copy_id", id), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "check_reservations_error",
			Message: "Не удалось проверить активные брони: " + err.Error(),
		}
	}
	if hasReservations {
		c.logger.Warn("copy has active reservations", zap.Int("copy_id", id))
		return nil, errors.BusinessError{
			Code:    "copy_has_reservations",
			Message: "Нельзя удалить копию, на которую есть активные брони",
		}
	}

	totalCopies, err := c.copyRepo.CountByBookID(ctx, conn, copyDel.BookID)
	if err != nil {
		c.logger.Error("failed to count copies by book", zap.Int("book_id", copyDel.BookID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "count_copies_error",
			Message: "Не удалось подсчитать копии книги: " + err.Error(),
		}
	}
	if totalCopies <= 1 {
		c.logger.Warn("cannot delete last copy of book", zap.Int("book_id", copyDel.BookID), zap.Int("total_copies", totalCopies))
		return nil, errors.BusinessError{
			Code:    "copy_is_last",
			Message: "Нельзя удалить последнюю копию книги",
		}
	}

	if err := c.copyRepo.Delete(ctx, conn, id); err != nil {
		c.logger.Error("failed to delete copy", zap.Int("copy_id", id), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrDeleteCopy",
			Message: "Не удалось удалить копию" + err.Error(),
		}
	}

	c.logger.Info("copy deleted successfully", zap.Int("copy_id", id), zap.Int("book_id", copyDel.BookID))
	return copyDel, nil
}

//	НЕ ЗАБЫТЬ
//
// добавить надо потом сюда лимит и оффсет ну и в другие файлы
func (c *CopyService) GetAvailableCopies(ctx context.Context, conn *pgx.Conn, bookID int) ([]domain.BookCopy, error) {
	c.logger.Debug("get available copies started", zap.Int("book_id", bookID))

	exists, err := c.bookRepo.Exists(ctx, conn, bookID)
	if err != nil {
		c.logger.Error("failed to check book existence", zap.Int("book_id", bookID), zap.Error(err))
		return nil, err
	}
	if !exists {
		c.logger.Warn("book not found for available copies", zap.Int("book_id", bookID))
		return nil, errors.NotFoundError{
			Entity: "Book",
			ID:     bookID,
		}
	}

	copes, err := c.copyRepo.GetAvailable(ctx, conn, bookID)
	if err != nil {
		c.logger.Error("failed to get available copies", zap.Int("book_id", bookID), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "BookCopyAvailable",
			ID:     bookID,
		}
	}

	c.logger.Debug("get available copies finished", zap.Int("book_id", bookID), zap.Int("available_count", len(copes)))
	return copes, nil
}

func (c *CopyService) CountAvailableCopies(ctx context.Context, conn *pgx.Conn, bookID int) (int, error) {
	c.logger.Debug("count available copies started", zap.Int("book_id", bookID))

	exists, err := c.bookRepo.Exists(ctx, conn, bookID)
	if err != nil {
		c.logger.Error("failed to check book existence", zap.Int("book_id", bookID), zap.Error(err))
		return 0, err
	}
	if !exists {
		c.logger.Warn("book not found for count available copies", zap.Int("book_id", bookID))
		return 0, errors.NotFoundError{
			Entity: "Book",
			ID:     bookID,
		}
	}
	count, err := c.copyRepo.CountAvailable(ctx, conn, bookID)
	if err != nil {
		c.logger.Error("failed to count available copies", zap.Int("book_id", bookID), zap.Error(err))
		return 0, errors.BusinessError{
			Code:    "ErrCountAvailable",
			Message: "Не удалось вернуть число доступных экземпляров:" + err.Error(),
		}
	}

	c.logger.Debug("count available copies finished", zap.Int("book_id", bookID), zap.Int("available_count", count))
	return count, nil
}

func (c *CopyService) UpdateCopyStatus(ctx context.Context, conn *pgx.Conn, id int, status string, readerID *int) error {
	c.logger.Info("update copy status started", zap.Int("copy_id", id), zap.String("status", status), zap.Intp("reader_id", readerID))

	_, err := c.copyRepo.GetByID(ctx, conn, id)
	if err != nil {
		c.logger.Warn("copy not found for status update", zap.Int("copy_id", id), zap.Error(err))
		return errors.NotFoundError{
			Entity: "BookCopy",
			ID:     id,
		}
	}

	if status != "available" && status != "borrowed" && status != "reserved" && status != "damaged" && status != "lost" {
		c.logger.Warn("invalid status", zap.Int("copy_id", id), zap.String("status", status))
		return errors.BusinessError{
			Code:    "invalid_status",
			Message: "Недопустимый статус",
		}
	}

	if status == "borrowed" {
		if readerID == nil {
			c.logger.Warn("reader_id required for borrowed status", zap.Int("copy_id", id))
			return errors.BusinessError{
				Code:    "reader_id_required",
				Message: "Для выдачи книги необходимо указать ID читателя",
			}
		}
		exists, err := c.readerRepo.Exists(ctx, conn, *readerID)
		if err != nil {
			c.logger.Error("failed to check reader existence", zap.Int("reader_id", *readerID), zap.Error(err))
			return err
		}
		if !exists {
			c.logger.Warn("reader not found", zap.Int("reader_id", *readerID))
			return errors.NotFoundError{
				Entity: "Reader",
				ID:     *readerID,
			}
		}
	}

	if err := c.copyRepo.UpdateStatus(ctx, conn, id, status); err != nil {
		c.logger.Error("failed to update copy status", zap.Int("copy_id", id), zap.String("status", status), zap.Error(err))
		return errors.BusinessError{
			Code:    "update_status_error",
			Message: "Не удалось обновить статус копии: " + err.Error(),
		}
	}

	if status == "available" {
		if err := c.copyRepo.ClearReaderAndBorrowed(ctx, conn, id); err != nil {
			c.logger.Error("failed to clear reader and borrowed data", zap.Int("copy_id", id), zap.Error(err))
			return errors.BusinessError{
				Code:    "clear_reader_error",
				Message: "Не удалось очистить данные читателя: " + err.Error(),
			}
		}
	}

	c.logger.Info("copy status updated successfully", zap.Int("copy_id", id), zap.String("status", status))
	return nil
}
