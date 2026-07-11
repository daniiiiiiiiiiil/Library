package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"

	"github.com/jackc/pgx/v5"
)

type CopyService struct {
	copyRepo        repository.BookCopyRepository
	bookRepo        repository.BookRepository
	readerRepo      repository.ReaderRepository
	reservationRepo repository.ReservationRepository
}

func (c *CopyService) CreateCopy(ctx context.Context, conn *pgx.Conn, copy domain.BookCopy) error {
	if err := copy.Validate(); err != nil {
		return errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации" + err.Error(),
		}
	}

	if copy.BookID > 0 {
		exists, err := c.bookRepo.Exists(ctx, conn, copy.BookID)
		if err != nil {
			return errors.BusinessError{
				Code:    "ErrBookExists",
				Message: "Не удалось проверить существование книги" + err.Error(),
			}
		}
		if !exists {
			return errors.NotFoundError{
				Entity: "BookNotFound",
				ID:     copy.BookID,
			}
		}
	}
	if err := c.copyRepo.Create(ctx, conn, copy); err != nil {
		return errors.BusinessError{
			Code:    "ErrCreateCopy",
			Message: "Не удалось создать копию" + err.Error(),
		}
	}

	return nil
}

func (c *CopyService) GetCopy(ctx context.Context, conn *pgx.Conn, id int) (*domain.BookCopy, error) {
	copyId, err := c.copyRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrGetCopy",
			Message: fmt.Sprintf("Не удалось получить копию %d по id", id) + err.Error(),
		}
	}
	return copyId, nil
}

func (c *CopyService) GetCopiesByBook(ctx context.Context, conn *pgx.Conn, bookID, limit, offset int) ([]domain.BookCopy, error) {
	limitOffset(limit, offset)
	copys, err := c.copyRepo.GetCopiesByBookID(ctx, conn, bookID, limit, offset)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrGetCopy",
			Message: "Не удалось получить все копии книг" + err.Error(),
		}
	}
	return copys, nil
}

func (c *CopyService) UpdateCopy(ctx context.Context, conn *pgx.Conn, copy *domain.BookCopy) error {
	if copy.ID > 0 {
		exists, err := c.copyRepo.ExistsCopy(ctx, conn, copy.ID)
		if err != nil {
			return errors.NotFoundError{
				Entity: "Copy",
				ID:     copy.ID,
			}
		}
		if !exists {
			return errors.NotFoundError{
				Entity: "Copy",
				ID:     copy.ID,
			}
		}
	}
	if err := copy.Validate(); err != nil {
		return errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации" + err.Error(),
		}
	}
	if copy.Status == "borrowed" {
		return errors.BusinessError{
			Code:    "ErrCopyBorrowedCannotUpdate",
			Message: "Копия выдана,нельзя обновить выданную копию",
		}
	}
	if err := c.copyRepo.Update(ctx, conn, *copy); err != nil {
		return errors.BusinessError{
			Code:    "ErrUpdateCopy",
			Message: "Не удалось обновить копию" + err.Error(),
		}
	}
	return nil
}

func (c *CopyService) DeleteCopy(ctx context.Context, conn *pgx.Conn, id int) (*domain.BookCopy, error) {
	copyDel, err := c.copyRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Copy",
			ID:     id,
		}
	}
	if copyDel.Status == "borrowed" {
		return nil, errors.BusinessError{
			Code:    "ErrCopyBorrowedCannotUpdate",
			Message: "Копия выдана,нельзя обновить выданную копию",
		}
	}

	hasReservations, err := c.reservationRepo.HasActiveForCopy(ctx, conn, id)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "check_reservations_error",
			Message: "Не удалось проверить активные брони: " + err.Error(),
		}
	}
	if hasReservations {
		return nil, errors.BusinessError{
			Code:    "copy_has_reservations",
			Message: "Нельзя удалить копию, на которую есть активные брони",
		}
	}

	totalCopies, err := c.copyRepo.CountByBookID(ctx, conn, copyDel.BookID)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "count_copies_error",
			Message: "Не удалось подсчитать копии книги: " + err.Error(),
		}
	}
	if totalCopies <= 1 {
		return nil, errors.BusinessError{
			Code:    "copy_is_last",
			Message: "Нельзя удалить последнюю копию книги",
		}
	}

	if err := c.copyRepo.Delete(ctx, conn, id); err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrDeleteCopy",
			Message: "Не удалось удалить копию" + err.Error(),
		}
	}
	return copyDel, nil
}

//	НЕ ЗАБЫТЬ
//
// добавить надо потом сюда лимит и оффсет ну и в другие файлы
func (c *CopyService) GetAvailableCopies(ctx context.Context, conn *pgx.Conn, bookID int) ([]domain.BookCopy, error) {
	exists, err := c.bookRepo.Exists(ctx, conn, bookID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NotFoundError{
			Entity: "Book",
			ID:     bookID,
		}
	}

	copes, err := c.copyRepo.GetAvailable(ctx, conn, bookID)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "BookCopyAvailable",
			ID:     bookID,
		}
	}
	return copes, nil
}

func (c *CopyService) CountAvailableCopies(ctx context.Context, conn *pgx.Conn, bookID int) (int, error) {
	exists, err := c.bookRepo.Exists(ctx, conn, bookID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, errors.NotFoundError{
			Entity: "Book",
			ID:     bookID,
		}
	}
	count, err := c.copyRepo.CountAvailable(ctx, conn, bookID)
	if err != nil {
		return 0, errors.BusinessError{
			Code:    "ErrCountAvailable",
			Message: "Не удалось вернуть число доступных экземпляров:" + err.Error(),
		}
	}
	return count, nil
}

func (c *CopyService) UpdateCopyStatus(ctx context.Context, conn *pgx.Conn, id int, status string, readerID *int) error {
	_, err := c.copyRepo.GetByID(ctx, conn, id)
	if err != nil {
		return errors.NotFoundError{
			Entity: "BookCopy",
			ID:     id,
		}
	}

	if status != "available" && status != "borrowed" && status != "reserved" && status != "damaged" && status != "lost" {
		return errors.BusinessError{
			Code:    "invalid_status",
			Message: "Недопустимый статус",
		}
	}

	if status == "borrowed" {
		if readerID == nil {
			return errors.BusinessError{
				Code:    "reader_id_required",
				Message: "Для выдачи книги необходимо указать ID читателя",
			}
		}
		exists, err := c.readerRepo.Exists(ctx, conn, *readerID)
		if err != nil {
			return err
		}
		if !exists {
			return errors.NotFoundError{
				Entity: "Reader",
				ID:     *readerID,
			}
		}
	}

	if err := c.copyRepo.UpdateStatus(ctx, conn, id, status); err != nil {
		return errors.BusinessError{
			Code:    "update_status_error",
			Message: "Не удалось обновить статус копии: " + err.Error(),
		}
	}

	if status == "available" {
		if err := c.copyRepo.ClearReaderAndBorrowed(ctx, conn, id); err != nil {
			return errors.BusinessError{
				Code:    "clear_reader_error",
				Message: "Не удалось очистить данные читателя: " + err.Error(),
			}
		}
	}

	return nil
}
