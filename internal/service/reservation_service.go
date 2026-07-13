package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type ReservationService struct {
	reservRepo repository.ReservationRepository
	copyRepo   repository.BookCopyRepository
	readerRepo repository.ReaderRepository
	bookRepo   repository.BookRepository
	txRepo     repository.TransactionRepository
	settings   repository.SettingRepository
}

func NewReservation(
	reservRepo repository.ReservationRepository,
	copyRepo repository.BookCopyRepository,
	readerRepo repository.ReaderRepository,
	bookRepo repository.BookRepository,
	txRepo repository.TransactionRepository,
	settings repository.SettingRepository,
) *ReservationService {
	return &ReservationService{
		reservRepo: reservRepo,
		copyRepo:   copyRepo,
		readerRepo: readerRepo,
		bookRepo:   bookRepo,
		txRepo:     txRepo,
		settings:   settings,
	}
}

func (s *ReservationService) ReserveBook(ctx context.Context, conn *pgx.Conn, copyID, readerID int) error {
	copy, err := s.copyRepo.GetByID(ctx, conn, copyID)
	if err != nil {
		return errors.NotFoundError{
			Entity: "BookCopy",
			ID:     copyID,
		}
	}

	if copy.Status != "available" {
		return errors.BusinessError{
			Code:    "ErrCopyNotAvailable",
			Message: fmt.Sprintf("Копия книги недоступна для бронирования (статус: %s)", copy.Status),
		}
	}

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
			Message: "Читатель заблокирован и не может бронировать книги",
		}
	}

	hasActive, err := s.reservRepo.IsBookReservedByOther(ctx, conn, copyID, readerID)
	if err != nil {
		return err
	}
	if hasActive {
		return errors.BusinessError{
			Code:    "ErrReservationAlreadyExists",
			Message: "У читателя уже есть активная бронь на эту книгу",
		}
	}

	reservedByOther, err := s.reservRepo.IsBookReservedByOther(ctx, conn, copyID, readerID)
	if err != nil {
		return err
	}
	if reservedByOther {
		return errors.BusinessError{
			Code:    "ErrReservedCopyByOther",
			Message: "Книга уже зарезервирована другим читателем",
		}
	}

	reservationDuration := 24 * time.Hour

	reservation := domain.NewReservation(
		copyID,
		readerID,
		time.Now().Add(reservationDuration),
	)

	if err := reservation.Validate(); err != nil {
		return errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации брони: " + err.Error(),
		}
	}

	if err := s.reservRepo.CreateReservation(ctx, conn, &reservation); err != nil {
		return errors.BusinessError{
			Code:    "ErrCreateReservation",
			Message: "Не удалось создать бронь: " + err.Error(),
		}
	}

	if err := s.copyRepo.UpdateStatus(ctx, conn, copyID, "reserved"); err != nil {
		return errors.BusinessError{
			Code:    "ErrReservationUpdatedStatus",
			Message: "Не удалось обновить статус копии: " + err.Error(),
		}
	}

	return nil
}

func (r *ReservationService) CancelReservation(ctx context.Context, conn *pgx.Conn, reservationID int) error {
	reserv, err := r.reservRepo.GetByID(ctx, conn, reservationID)
	if err != nil {
		return errors.NotFoundError{
			Entity: "Reservation",
			ID:     reservationID,
		}
	}
	if reserv.Status != "active" {
		return errors.BusinessError{
			Code:    "ErrReservationNotActive",
			Message: "Резерв не активен" + err.Error(),
		}
	}
	if err := r.reservRepo.UpdateStatus(ctx, conn, reservationID, "cancelled"); err != nil {
		return errors.BusinessError{
			Code:    "ErrReservationCancelled",
			Message: "Не удалось отменить резервацию: " + err.Error(),
		}
	}
	if err := r.copyRepo.UpdateStatus(ctx, conn, reserv.CopyID, "available"); err != nil {
		return errors.BusinessError{
			Code:    "ErrCopyUpdatedStatus",
			Message: "Не удалось поменять статус копии книги на available" + err.Error(),
		}
	}
	return err
}

func (r *ReservationService) GetReservation(ctx context.Context, conn *pgx.Conn, reservationID int) (*domain.Reservation, error) {
	if reservationID <= 0 {
		return nil, errors.ValidationError{
			Field:   "ReservationID",
			Message: "ID меньше 0 или равен 0",
		}
	}
	reserv, err := r.reservRepo.GetByID(ctx, conn, reservationID)
	if err != nil {
		return nil, errors.NotFoundError{}
	}
	return reserv, nil
}

func (r *ReservationService) GetActiveReaderReservations(ctx context.Context, conn *pgx.Conn, readerID int, limit int, offset int) ([]domain.Reservation, int, error) {
	limitOffset(limit, offset)
	exists, err := r.readerRepo.Exists(ctx, conn, readerID)
	if err != nil {
		return nil, 0, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}
	if !exists {
		return nil, 0, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}
	listActiveReaders, err := r.reservRepo.GetActiveByReader(ctx, conn, readerID, limit, offset)
	if err != nil {
		return nil, 0, errors.NotFoundError{
			Entity: "ReaderActive",
			ID:     readerID,
		}
	}
	return listActiveReaders, len(listActiveReaders), nil
}

func (r *ReservationService) GetActiveCopyReservations(ctx context.Context, conn *pgx.Conn, copyID int, limit, offset int) ([]domain.Reservation, int, error) {
	exists, err := r.copyRepo.ExistsCopy(ctx, conn, copyID)
	if err != nil {
		return nil, 0, errors.NotFoundError{
			Entity: "Copy",
			ID:     copyID,
		}
	}
	if !exists {
		return nil, 0, errors.NotFoundError{
			Entity: "Copy",
			ID:     copyID,
		}
	}
	activeCopy, err := r.reservRepo.GetActiveByCopy(ctx, conn, copyID, limit, offset)
	if err != nil {
		return nil, 0, errors.NotFoundError{
			Entity: "CopyActive",
			ID:     copyID,
		}
	}
	return activeCopy, len(activeCopy), nil
}

func (r *ReservationService) ProcessExpiredReservations(ctx context.Context, conn *pgx.Conn, limit, offset int) error {
	limitOffset(limit, offset)

	expiredReservations, err := r.reservRepo.GetExpired(ctx, conn, 100, 0)
	if err != nil {
		return errors.BusinessError{
			Code:    "get_expired_error",
			Message: "Не удалось получить просроченные брони: " + err.Error(),
		}
	}
	for _, reservation := range expiredReservations {
		if err := r.reservRepo.UpdateStatus(ctx, conn, reservation.ID, "expired"); err != nil {
			return errors.BusinessError{
				Code:    "ErrReservationUpdateStatus",
				Message: "Не удалось обновить статус резервации на expired: " + err.Error(),
			}
		}
		if err := r.copyRepo.UpdateStatus(ctx, conn, reservation.ID, "available"); err != nil {
			return errors.BusinessError{
				Code:    "ErrCopyUpdateStatus",
				Message: "Не удалось обновить статус у копии на available: " + err.Error(),
			}
		}
	}
	return nil
}

func (r *ReservationService) CanReserve(ctx context.Context, conn *pgx.Conn, reservationID, copyID, readerID int) (bool, error) {
	exists, err := r.copyRepo.ExistsCopy(ctx, conn, copyID)
	if err != nil {
		return false, errors.NotFoundError{
			Entity: "Copy",
			ID:     copyID,
		}
	}
	if !exists {
		return false, errors.NotFoundError{
			Entity: "Copy",
			ID:     copyID,
		}
	}
	_, err = r.copyRepo.GetAvailable(ctx, conn, copyID)
	if err != nil {
		return false, errors.NotFoundError{
			Entity: "GetAvailableBook",
			ID:     copyID,
		}
	}
	exists, err = r.readerRepo.Exists(ctx, conn, readerID)
	if err != nil {
		return false, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}
	if !exists {
		return false, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}
	reader, err := r.readerRepo.GetByID(ctx, conn, readerID)
	if err != nil {
		return false, errors.NotFoundError{Entity: "Reader", ID: readerID}
	}

	if reader.IsBlocked() {
		return false, errors.BusinessError{
			Code:    "ErrReaderIsBlocked",
			Message: "Читатель заблокирован",
		}
	}
	_, err = r.reservRepo.HasActiveForCopy(ctx, conn, copyID)
	if err != nil {
		return false, errors.BusinessError{
			Code:    "ErrHasActiveForCopy",
			Message: "Уже есть активная бронь на эту книгу" + err.Error(),
		}
	}
	return true, nil
}
