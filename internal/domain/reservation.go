package domain

import (
	"errors"
	"fmt"
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

func (r Reservation) CalculateTimeRemaining() string {
	now := time.Now()

	if now.After(r.ExpiresAt) {
		return "истекла"
	}

	duration := r.ExpiresAt.Sub(now)

	if duration < time.Minute {
		return "менее минуты"
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return formatMinutes(minutes)
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		return formatHoursMinutes(hours, minutes)
	}

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	return formatDaysHours(days, hours)
}

func formatMinutes(minutes int) string {
	lastDigit := minutes % 10
	if minutes >= 11 && minutes <= 14 {
		return fmt.Sprintf("%d минут", minutes)
	}
	switch lastDigit {
	case 1:
		return fmt.Sprintf("%d минута", minutes)
	case 2, 3, 4:
		return fmt.Sprintf("%d минуты", minutes)
	default:
		return fmt.Sprintf("%d минут", minutes)
	}
}

func formatHoursMinutes(hours, minutes int) string {
	hoursStr := "часов"
	lastDigit := hours % 10
	if hours == 1 || (hours%10 == 1 && hours != 11) {
		hoursStr = "час"
	} else if hours >= 2 && hours <= 4 || (lastDigit >= 2 && lastDigit <= 4 && hours != 12 && hours != 13 && hours != 14) {
		hoursStr = "часа"
	}

	if minutes == 0 {
		return fmt.Sprintf("%d %s", hours, hoursStr)
	}
	return fmt.Sprintf("%d %s %d минут", hours, hoursStr, minutes)
}

func formatDaysHours(days, hours int) string {
	daysStr := "дней"
	lastDigit := days % 10
	if days == 1 || (days%10 == 1 && days != 11) {
		daysStr = "день"
	} else if days >= 2 && days <= 4 || (lastDigit >= 2 && lastDigit <= 4 && days != 12 && days != 13 && days != 14) {
		daysStr = "дня"
	}

	if hours == 0 {
		return fmt.Sprintf("%d %s", days, daysStr)
	}
	return fmt.Sprintf("%d %s %d часов", days, daysStr, hours)
}

type ProcessExpiredRequest struct {
	Limit int `json:"limit"`
}

func (r *ProcessExpiredRequest) Validate() error {
	if r.Limit <= 0 {
		r.Limit = 100
	}
	return nil
}

type ProcessExpiredResponse struct {
	ProcessedCount int `json:"processed_count"`
}
