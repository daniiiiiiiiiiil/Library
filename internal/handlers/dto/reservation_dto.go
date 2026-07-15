package dto

import (
	"library/internal/domain"
	"library/pkg/errors"
	"library/pkg/pagination"
	"time"
)

type CreateReservationRequest struct {
	Copy_id int `json:"CopyID"`
}

func (r *CreateReservationRequest) Validate() error {
	if r.Copy_id < 0 {
		return errors.ValidationError{
			Field:   "CopyID",
			Message: "ID копии не может быть меньше нуля",
		}
	}
	return nil
}

func (r *CreateReservationRequest) ToDomain(readerID int) domain.Reservation {
	return domain.NewReservation(
		r.Copy_id,
		readerID,
		time.Now().Add(24*time.Hour),
	)
}

type CancelReservationRequest struct {
	ReservationID int `json:"reservation_id"`
}

type ReservationResponse struct {
	ID            int       `json:"ID"`
	CopyID        int       `json:"CopyID"`
	BookTitle     string    `json:"BookTitle"`
	ReaderID      int       `json:"reader_id"`
	ReaderName    string    `json:"ReaderName"`
	ReserverAt    time.Time `json:"reserved_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	Status        string    `json:"status"`
	IsExpired     bool      `json:"is_expired"`
	TimeRemaining string    `json:"time_remaining"`
}

func (r *ReservationResponse) FromDomain(d domain.Reservation, bookTitle, readerName string) ReservationResponse {
	return ReservationResponse{
		ID:            d.ID,
		CopyID:        d.CopyID,
		BookTitle:     bookTitle,
		ReaderID:      d.ReaderID,
		ReaderName:    readerName,
		ReserverAt:    d.ReservedAt,
		ExpiresAt:     d.ExpiresAt,
		Status:        string(d.Status),
		IsExpired:     d.IsExpired(),
		TimeRemaining: d.CalculateTimeRemaining(),
	}
}

type ReservationListResponse struct {
	reservations []ReservationResponse
	pagination   pagination.Pagination
}
