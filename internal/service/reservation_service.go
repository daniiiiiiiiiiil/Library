package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ReservationService struct {
	reservRepo repository.ReservationRepository
	copyRepo   repository.BookCopyRepository
	readerRepo repository.ReaderRepository
	bookRepo   repository.BookRepository
	txRepo     repository.TransactionRepository
	settings   repository.SettingRepository
	logger     *zap.Logger
}

func NewReservation(
	reservRepo repository.ReservationRepository,
	copyRepo repository.BookCopyRepository,
	readerRepo repository.ReaderRepository,
	bookRepo repository.BookRepository,
	txRepo repository.TransactionRepository,
	settings repository.SettingRepository,
	logger *zap.Logger,
) *ReservationService {
	return &ReservationService{
		reservRepo: reservRepo,
		copyRepo:   copyRepo,
		readerRepo: readerRepo,
		bookRepo:   bookRepo,
		txRepo:     txRepo,
		settings:   settings,
		logger:     logger,
	}
}

func (s *ReservationService) ReserveBook(ctx context.Context, conn *pgx.Conn, copyID, readerID int) error {
	s.logger.Info("reserve book started", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID))

	copy, err := s.copyRepo.GetByID(ctx, conn, copyID)
	if err != nil {
		s.logger.Error("failed to get copy", zap.Int("copy_id", copyID), zap.Error(err))
		return errors.NotFoundError{
			Entity: "BookCopy",
			ID:     copyID,
		}
	}

	if copy.Status != "available" {
		s.logger.Warn("copy not available for reservation", zap.Int("copy_id", copyID), zap.String("status", copy.Status))
		return errors.BusinessError{
			Code:    "ErrCopyNotAvailable",
			Message: fmt.Sprintf("Копия книги недоступна для бронирования (статус: %s)", copy.Status),
		}
	}

	reader, err := s.readerRepo.GetByID(ctx, conn, readerID)
	if err != nil {
		s.logger.Error("failed to get reader", zap.Int("reader_id", readerID), zap.Error(err))
		return errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}
	if reader.Status == "blocked" {
		s.logger.Warn("reader is blocked", zap.Int("reader_id", readerID))
		return errors.BusinessError{
			Code:    "ErrReaderBlocked",
			Message: "Читатель заблокирован и не может бронировать книги",
		}
	}

	hasActive, err := s.reservRepo.IsBookReservedByOther(ctx, conn, copyID, readerID)
	if err != nil {
		s.logger.Error("failed to check if book reserved by other", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID), zap.Error(err))
		return err
	}
	if hasActive {
		s.logger.Warn("reader already has active reservation for this copy", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID))
		return errors.BusinessError{
			Code:    "ErrReservationAlreadyExists",
			Message: "У читателя уже есть активная бронь на эту книгу",
		}
	}

	reservedByOther, err := s.reservRepo.IsBookReservedByOther(ctx, conn, copyID, readerID)
	if err != nil {
		s.logger.Error("failed to check if book reserved by other", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID), zap.Error(err))
		return err
	}
	if reservedByOther {
		s.logger.Warn("book already reserved by other reader", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID))
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
		s.logger.Warn("reservation validation failed", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID), zap.Error(err))
		return errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации брони: " + err.Error(),
		}
	}

	if err := s.reservRepo.CreateReservation(ctx, conn, &reservation); err != nil {
		s.logger.Error("failed to create reservation", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrCreateReservation",
			Message: "Не удалось создать бронь: " + err.Error(),
		}
	}

	if err := s.copyRepo.UpdateStatus(ctx, conn, copyID, "reserved"); err != nil {
		s.logger.Error("failed to update copy status to reserved", zap.Int("copy_id", copyID), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrReservationUpdatedStatus",
			Message: "Не удалось обновить статус копии: " + err.Error(),
		}
	}

	s.logger.Info("book reserved successfully", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID), zap.Time("expires_at", reservation.ExpiresAt))
	return nil
}

func (r *ReservationService) CancelReservation(ctx context.Context, conn *pgx.Conn, reservationID int) error {
	r.logger.Info("cancel reservation started", zap.Int("reservation_id", reservationID))

	reserv, err := r.reservRepo.GetByID(ctx, conn, reservationID)
	if err != nil {
		r.logger.Error("failed to get reservation", zap.Int("reservation_id", reservationID), zap.Error(err))
		return errors.NotFoundError{
			Entity: "Reservation",
			ID:     reservationID,
		}
	}

	if reserv.Status != "active" {
		r.logger.Warn("reservation is not active", zap.Int("reservation_id", reservationID), zap.String("status", ""))
		return errors.BusinessError{
			Code:    "ErrReservationNotActive",
			Message: "Резерв не активен" + err.Error(),
		}
	}

	if err := r.reservRepo.UpdateStatus(ctx, conn, reservationID, "cancelled"); err != nil {
		r.logger.Error("failed to update reservation status to cancelled", zap.Int("reservation_id", reservationID), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrReservationCancelled",
			Message: "Не удалось отменить резервацию: " + err.Error(),
		}
	}

	if err := r.copyRepo.UpdateStatus(ctx, conn, reserv.CopyID, "available"); err != nil {
		r.logger.Error("failed to update copy status to available", zap.Int("copy_id", reserv.CopyID), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrCopyUpdatedStatus",
			Message: "Не удалось поменять статус копии книги на available" + err.Error(),
		}
	}

	r.logger.Info("reservation cancelled successfully", zap.Int("reservation_id", reservationID), zap.Int("copy_id", reserv.CopyID))
	return nil
}

func (r *ReservationService) GetReservation(ctx context.Context, conn *pgx.Conn, reservationID int) (*domain.Reservation, error) {
	r.logger.Debug("get reservation started", zap.Int("reservation_id", reservationID))

	if reservationID <= 0 {
		r.logger.Warn("invalid reservation id", zap.Int("reservation_id", reservationID))
		return nil, errors.ValidationError{
			Field:   "ReservationID",
			Message: "ID меньше 0 или равен 0",
		}
	}

	reserv, err := r.reservRepo.GetByID(ctx, conn, reservationID)
	if err != nil {
		r.logger.Warn("reservation not found", zap.Int("reservation_id", reservationID), zap.Error(err))
		return nil, errors.NotFoundError{}
	}

	r.logger.Debug("get reservation finished", zap.Int("reservation_id", reserv.ID))
	return reserv, nil
}

func (r *ReservationService) GetActiveReaderReservations(ctx context.Context, conn *pgx.Conn, readerID int, limit int, offset int) ([]domain.Reservation, int, error) {
	limit, offset = limitOffset(limit, offset)
	r.logger.Debug("get active reader reservations started", zap.Int("reader_id", readerID), zap.Int("limit", limit), zap.Int("offset", offset))

	exists, err := r.readerRepo.Exists(ctx, conn, readerID)
	if err != nil {
		r.logger.Error("failed to check reader existence", zap.Int("reader_id", readerID), zap.Error(err))
		return nil, 0, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}
	if !exists {
		r.logger.Warn("reader not found", zap.Int("reader_id", readerID))
		return nil, 0, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}

	listActiveReaders, err := r.reservRepo.GetActiveByReader(ctx, conn, readerID, limit, offset)
	if err != nil {
		r.logger.Error("failed to get active reader reservations", zap.Int("reader_id", readerID), zap.Error(err))
		return nil, 0, errors.NotFoundError{
			Entity: "ReaderActive",
			ID:     readerID,
		}
	}

	r.logger.Debug("get active reader reservations finished", zap.Int("reader_id", readerID), zap.Int("returned", len(listActiveReaders)))
	return listActiveReaders, len(listActiveReaders), nil
}

func (r *ReservationService) GetActiveCopyReservations(ctx context.Context, conn *pgx.Conn, copyID int, limit, offset int) ([]domain.Reservation, int, error) {
	limit, offset = limitOffset(limit, offset)
	r.logger.Debug("get active copy reservations started", zap.Int("copy_id", copyID), zap.Int("limit", limit), zap.Int("offset", offset))

	exists, err := r.copyRepo.ExistsCopy(ctx, conn, copyID)
	if err != nil {
		r.logger.Error("failed to check copy existence", zap.Int("copy_id", copyID), zap.Error(err))
		return nil, 0, errors.NotFoundError{
			Entity: "Copy",
			ID:     copyID,
		}
	}
	if !exists {
		r.logger.Warn("copy not found", zap.Int("copy_id", copyID))
		return nil, 0, errors.NotFoundError{
			Entity: "Copy",
			ID:     copyID,
		}
	}

	activeCopy, err := r.reservRepo.GetActiveByCopy(ctx, conn, copyID, limit, offset)
	if err != nil {
		r.logger.Error("failed to get active copy reservations", zap.Int("copy_id", copyID), zap.Error(err))
		return nil, 0, errors.NotFoundError{
			Entity: "CopyActive",
			ID:     copyID,
		}
	}

	r.logger.Debug("get active copy reservations finished", zap.Int("copy_id", copyID), zap.Int("returned", len(activeCopy)))
	return activeCopy, len(activeCopy), nil
}

func (r *ReservationService) ProcessExpiredReservations(ctx context.Context, conn *pgx.Conn, limit, offset int) error {
	limit, offset = limitOffset(limit, offset)
	r.logger.Info("process expired reservations started", zap.Int("limit", limit), zap.Int("offset", offset))

	expiredReservations, err := r.reservRepo.GetExpired(ctx, conn, 100, 0)
	if err != nil {
		r.logger.Error("failed to get expired reservations", zap.Error(err))
		return errors.BusinessError{
			Code:    "get_expired_error",
			Message: "Не удалось получить просроченные брони: " + err.Error(),
		}
	}

	if len(expiredReservations) == 0 {
		r.logger.Info("no expired reservations found")
		return nil
	}

	r.logger.Info("processing expired reservations", zap.Int("count", len(expiredReservations)))

	for _, reservation := range expiredReservations {
		if err := r.reservRepo.UpdateStatus(ctx, conn, reservation.ID, "expired"); err != nil {
			r.logger.Error("failed to update reservation status to expired", zap.Int("reservation_id", reservation.ID), zap.Error(err))
			return errors.BusinessError{
				Code:    "ErrReservationUpdateStatus",
				Message: "Не удалось обновить статус резервации на expired: " + err.Error(),
			}
		}

		if err := r.copyRepo.UpdateStatus(ctx, conn, reservation.CopyID, "available"); err != nil {
			r.logger.Error("failed to update copy status to available", zap.Int("copy_id", reservation.CopyID), zap.Error(err))
			return errors.BusinessError{
				Code:    "ErrCopyUpdateStatus",
				Message: "Не удалось обновить статус у копии на available: " + err.Error(),
			}
		}

		r.logger.Debug("expired reservation processed", zap.Int("reservation_id", reservation.ID), zap.Int("copy_id", reservation.CopyID))
	}

	r.logger.Info("process expired reservations completed", zap.Int("processed", len(expiredReservations)))
	return nil
}

func (r *ReservationService) CanReserve(ctx context.Context, conn *pgx.Conn, reservationID, copyID, readerID int) (bool, error) {
	r.logger.Debug("can reserve check started", zap.Int("reservation_id", reservationID), zap.Int("copy_id", copyID), zap.Int("reader_id", readerID))

	exists, err := r.copyRepo.ExistsCopy(ctx, conn, copyID)
	if err != nil {
		r.logger.Error("failed to check copy existence", zap.Int("copy_id", copyID), zap.Error(err))
		return false, errors.NotFoundError{
			Entity: "Copy",
			ID:     copyID,
		}
	}
	if !exists {
		r.logger.Warn("copy not found", zap.Int("copy_id", copyID))
		return false, errors.NotFoundError{
			Entity: "Copy",
			ID:     copyID,
		}
	}

	_, err = r.copyRepo.GetAvailable(ctx, conn, copyID)
	if err != nil {
		r.logger.Error("failed to get available copies", zap.Int("copy_id", copyID), zap.Error(err))
		return false, errors.NotFoundError{
			Entity: "GetAvailableBook",
			ID:     copyID,
		}
	}

	exists, err = r.readerRepo.Exists(ctx, conn, readerID)
	if err != nil {
		r.logger.Error("failed to check reader existence", zap.Int("reader_id", readerID), zap.Error(err))
		return false, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}
	if !exists {
		r.logger.Warn("reader not found", zap.Int("reader_id", readerID))
		return false, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}

	reader, err := r.readerRepo.GetByID(ctx, conn, readerID)
	if err != nil {
		r.logger.Error("failed to get reader", zap.Int("reader_id", readerID), zap.Error(err))
		return false, errors.NotFoundError{Entity: "Reader", ID: readerID}
	}

	if reader.IsBlocked() {
		r.logger.Warn("reader is blocked", zap.Int("reader_id", readerID))
		return false, errors.BusinessError{
			Code:    "ErrReaderIsBlocked",
			Message: "Читатель заблокирован",
		}
	}

	hasActive, err := r.reservRepo.HasActiveForCopy(ctx, conn, copyID)
	if err != nil {
		r.logger.Error("failed to check active reservations for copy", zap.Int("copy_id", copyID), zap.Error(err))
		return false, errors.BusinessError{
			Code:    "ErrHasActiveForCopy",
			Message: "Уже есть активная бронь на эту книгу" + err.Error(),
		}
	}
	if hasActive {
		r.logger.Warn("copy already has active reservation", zap.Int("copy_id", copyID))
		return false, errors.BusinessError{
			Code:    "ErrHasActiveForCopy",
			Message: "Уже есть активная бронь на эту книгу",
		}
	}

	r.logger.Debug("can reserve check passed", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID))
	return true, nil
}
