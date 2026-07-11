package domain

import (
	"errors"
	"time"
)

type ReservationStatus string

const (
	StatusActive    ReservationStatus = "active"
	StatusCompleted ReservationStatus = "completed"
	StatusExpired   ReservationStatus = "expired"
	StatusCancelled ReservationStatus = "cancelled"
)

type Reservation struct {
	ID         int
	CopyID     int
	ReaderID   int
	ReservedAt time.Time
	ExpiresAt  time.Time
	Status     ReservationStatus
}

func NewReservation(copyID, readerID int, expiresAt time.Time) Reservation {
	return Reservation{
		CopyID:     copyID,
		ReaderID:   readerID,
		ReservedAt: time.Now(),
		ExpiresAt:  expiresAt,
		Status:     StatusActive,
	}
}

func (r Reservation) Validate() error {
	if r.ReaderID <= 0 {
		return errors.New("ID читателя не может быть равен 0")
	}
	if r.ExpiresAt.Before(time.Now()) {
		return errors.New("Время должно быть больше чем сейчас")
	}
	return nil
}

func (r Reservation) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

func (r Reservation) IsActive() bool {
	return r.Status == StatusActive && !r.IsExpired()
}

func (r *Reservation) Expire() {
	r.Status = StatusExpired
}

func (r *Reservation) Cancel() {
	r.Status = StatusCancelled
}

func (r *Reservation) Complete() {
	r.Status = StatusCompleted
}
