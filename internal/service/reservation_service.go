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

type ReservationResult struct {
	Reservation domain.Reservation
	BookTitle   string
	ReaderName  string
}

func NewReservationResult() *ReservationResult {
	return &ReservationResult{
		Reservation: domain.Reservation{},
		BookTitle:   "",
		ReaderName:  "",
	}
}

func (s *ReservationService) ReserveBook(ctx context.Context, conn *pgx.Conn, copyID, readerID int) (*ReservationResult, error) {
	s.logger.Info("reserve book started", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID))

	copy, err := s.copyRepo.GetByID(ctx, conn, copyID)
	if err != nil {
		s.logger.Error("failed to get copy", zap.Int("copy_id", copyID), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "BookCopy",
			ID:     copyID,
		}
	}

	if copy.Status != "available" {
		s.logger.Warn("copy not available for reservation", zap.Int("copy_id", copyID), zap.String("status", copy.Status))
		return nil, errors.BusinessError{
			Code:    "ErrCopyNotAvailable",
			Message: fmt.Sprintf("Копия книги недоступна для бронирования (статус: %s)", copy.Status),
		}
	}

	reader, err := s.readerRepo.GetByID(ctx, conn, readerID)
	if err != nil {
		s.logger.Error("failed to get reader", zap.Int("reader_id", readerID), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}

	if reader.Status == "blocked" {
		s.logger.Warn("reader is blocked", zap.Int("reader_id", readerID))
		return nil, errors.BusinessError{
			Code:    "ErrReaderBlocked",
			Message: "Читатель заблокирован и не может бронировать книги",
		}
	}

	reservedByOther, err := s.reservRepo.IsBookReservedByOther(ctx, conn, copyID, readerID)
	if err != nil {
		s.logger.Error("failed to check if book reserved by other", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID), zap.Error(err))
		return nil, err
	}
	if reservedByOther {
		s.logger.Warn("book already reserved by other reader", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID))
		return nil, errors.BusinessError{
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
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации брони: " + err.Error(),
		}
	}

	created, err := s.reservRepo.CreateReservation(ctx, conn, &reservation)
	if err != nil {
		s.logger.Error("failed to create reservation", zap.Int("copy_id", copyID), zap.Int("reader_id", readerID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrCreateReservation",
			Message: "Не удалось создать бронь: " + err.Error(),
		}
	}

	_, err = s.copyRepo.UpdateStatus(ctx, conn, copyID, "reserved")
	if err != nil {
		s.logger.Error("failed to update copy status to reserved", zap.Int("copy_id", copyID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrReservationUpdatedStatus",
			Message: "Не удалось обновить статус копии: " + err.Error(),
		}
	}

	book, err := s.bookRepo.GetByID(ctx, conn, copy.BookID)
	if err != nil {
		s.logger.Error("failed to get book for reservation response", zap.Int("book_id", copy.BookID), zap.Error(err))
		book.Title = "" // Не фатально, возвращаем пустую строку
	}

	s.logger.Info("book reserved successfully",
		zap.Int("copy_id", copyID),
		zap.Int("reader_id", readerID),
		zap.Time("expires_at", reservation.ExpiresAt),
	)

	return &ReservationResult{
		Reservation: *created,
		BookTitle:   book.Title,
		ReaderName:  reader.Name,
	}, nil
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

	_, err = r.copyRepo.UpdateStatus(ctx, conn, reserv.CopyID, "available")
	if err != nil {
		r.logger.Error("failed to update copy status to available", zap.Int("copy_id", reserv.CopyID), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrCopyUpdatedStatus",
			Message: "Не удалось поменять статус копии книги на available" + err.Error(),
		}
	}

	r.logger.Info("reservation cancelled successfully", zap.Int("reservation_id", reservationID), zap.Int("copy_id", reserv.CopyID))
	return nil
}

func (s *ReservationService) GetReservation(ctx context.Context, conn *pgx.Conn, reservationID int) (*ReservationResult, error) {
	s.logger.Debug("get reservation started", zap.Int("reservation_id", reservationID))

	reservation, err := s.reservRepo.GetByID(ctx, conn, reservationID)
	if err != nil {
		s.logger.Warn("reservation not found", zap.Int("reservation_id", reservationID), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Reservation",
			ID:     reservationID,
		}
	}

	copy, err := s.copyRepo.GetByID(ctx, conn, reservation.CopyID)
	if err != nil {
		s.logger.Error("failed to get copy for reservation", zap.Int("copy_id", reservation.CopyID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrGetCopy",
			Message: "Не удалось получить копию книги: " + err.Error(),
		}
	}

	book, err := s.bookRepo.GetByID(ctx, conn, copy.BookID)
	if err != nil {
		s.logger.Error("failed to get book for reservation", zap.Int("book_id", copy.BookID), zap.Error(err))
		book.Title = ""
	}

	reader, err := s.readerRepo.GetByID(ctx, conn, reservation.ReaderID)
	if err != nil {
		s.logger.Error("failed to get reader for reservation", zap.Int("reader_id", reservation.ReaderID), zap.Error(err))
		reader.Name = ""
	}

	s.logger.Debug("get reservation finished", zap.Int("reservation_id", reservationID))

	return &ReservationResult{
		Reservation: *reservation,
		BookTitle:   book.Title,
		ReaderName:  reader.Name,
	}, nil
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

	listActiveReaders, total, err := r.reservRepo.GetActiveByReader(ctx, conn, readerID, limit, offset)
	if err != nil {
		r.logger.Error("failed to get active reader reservations", zap.Int("reader_id", readerID), zap.Error(err))
		return nil, 0, errors.NotFoundError{
			Entity: "ReaderActive",
			ID:     readerID,
		}
	}

	r.logger.Debug("get active reader reservations finished", zap.Int("reader_id", readerID), zap.Int("returned", total))
	return listActiveReaders, total, nil
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

func (s *ReservationService) ProcessExpiredReservations(ctx context.Context, conn *pgx.Conn, limit int) (int, error) {
	s.logger.Info("processing expired reservations started", zap.Int("limit", limit))

	expiredReservations, err := s.reservRepo.GetExpired(ctx, conn, limit, 0)
	if err != nil {
		s.logger.Error("failed to get expired reservations", zap.Error(err))
		return 0, errors.BusinessError{
			Code:    "ErrGetExpired",
			Message: "Не удалось получить просроченные брони: " + err.Error(),
		}
	}

	if len(expiredReservations) == 0 {
		s.logger.Info("no expired reservations found")
		return 0, nil
	}

	processedCount := 0
	for _, res := range expiredReservations {
		if err := s.reservRepo.UpdateStatus(ctx, conn, res.ID, "expired"); err != nil {
			s.logger.Error("failed to update reservation status to expired", zap.Int("reservation_id", res.ID), zap.Error(err))
			continue
		}

		_, err := s.copyRepo.UpdateStatus(ctx, conn, res.CopyID, "available")
		if err != nil {
			s.logger.Error("failed to update copy status to available", zap.Int("copy_id", res.CopyID), zap.Error(err))
			continue
		}

		processedCount++
		s.logger.Info("expired reservation processed",
			zap.Int("reservation_id", res.ID),
			zap.Int("copy_id", res.CopyID),
			zap.Int("reader_id", res.ReaderID),
		)
	}

	s.logger.Info("processing expired reservations finished", zap.Int("processed_count", processedCount))
	return processedCount, nil
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

type ReservationWithDetails struct {
	Reservation domain.Reservation
	BookTitle   string
	ReaderName  string
}

func (s *ReservationService) GetActiveReaderReservationsWithDetails(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]ReservationWithDetails, int, error) {
	reservations, total, err := s.reservRepo.GetActiveByReader(ctx, conn, readerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	result := make([]ReservationWithDetails, 0, len(reservations))
	for _, res := range reservations {
		copy, err := s.copyRepo.GetByID(ctx, conn, res.CopyID)
		if err != nil {
			continue
		}
		book, err := s.bookRepo.GetByID(ctx, conn, copy.BookID)
		if err != nil {
			continue
		}
		reader, err := s.readerRepo.GetByID(ctx, conn, res.ReaderID)
		if err != nil {
			continue
		}

		result = append(result, ReservationWithDetails{
			Reservation: res,
			BookTitle:   book.Title,
			ReaderName:  reader.Name,
		})
	}

	return result, total, nil
}

func (s *ReservationService) GetActiveReservationsByCopy(ctx context.Context, conn *pgx.Conn, copyID, limit, offset int) ([]domain.Reservation, int, error) {
	s.logger.Debug("get active reservations by copy started", zap.Int("copy_id", copyID))

	reservations, err := s.reservRepo.GetActiveByCopy(ctx, conn, copyID, limit, offset)
	if err != nil {
		s.logger.Error("failed to get active reservations by copy", zap.Int("copy_id", copyID), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrGetActiveReservations",
			Message: "Не удалось получить активные брони: " + err.Error(),
		}
	}

	total := len(reservations)
	s.logger.Debug("get active reservations by copy finished", zap.Int("copy_id", copyID), zap.Int("count", total))
	return reservations, total, nil
}
