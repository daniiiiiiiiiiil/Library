package dto

import (
	"library/internal/domain"
	"library/pkg/errors"
	"library/pkg/pagination"
	"strings"
	"time"
)

type CreateReviewRequest struct {
	BookID  int     `json:"book_id"`
	Rating  float64 `json:"Rating"`
	Comment string  `json:"Comment"`
}

func (r *CreateReviewRequest) Validate() error {
	var errs errors.ValidationErrors
	if r.BookID < 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "book_id",
			Message: "ID книги не может быть пустым",
		})
	}
	if r.Rating < 0 || r.Rating > 11 {
		errs = append(errs, errors.ValidationError{
			Field:   "Rating",
			Message: "Рейтинг должен быть от 1 до 10",
		})
	}
	if strings.TrimSpace(r.Comment) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "Comment",
			Message: "Комментарий не может быть пустым",
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *CreateReviewRequest) ToDomain(readerID int) domain.Review {
	return domain.NewReview(
		0,
		r.BookID,
		readerID,
		r.Rating,
		r.Comment,
		time.Now(),
		nil,
	)
}

type UpdateReviewRequest struct {
	Rating  *float64 `json:"Rating"`
	Comment *string  `json:"Comment"`
}

func (r *UpdateReviewRequest) Validate() error {
	var errs errors.ValidationErrors
	if r.Rating != nil && (*r.Rating < 0 || *r.Rating > 10) {
		errs = append(errs, errors.ValidationError{
			Field:   "Rating",
			Message: "Рейтинг должен быть от 1 до 10",
		})
	}
	if r.Comment != nil && strings.TrimSpace(*r.Comment) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "Comment",
			Message: "Комментарий не должен быть пустым",
		})
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *UpdateReviewRequest) ToDomain(reviewID int, review domain.Review) domain.Review {
	rating := review.Rating
	if r.Rating != nil {
		rating = *r.Rating
	}
	comment := review.Comment
	if r.Comment != nil {
		comment = *r.Comment
	}
	return domain.NewReview(
		reviewID,
		review.BookID,
		review.ReaderID,
		rating,
		comment,
		review.CreatedAt,
		review.UpdatedAt,
	)
}

type ReviewResponse struct {
	ID         int        `json:"ID"`
	BookID     int        `json:"book_id"`
	BookTitle  string     `json:"BookTitle"`
	ReaderID   int        `json:"reader_id"`
	ReaderName string     `json:"ReaderName"`
	Rating     float64    `json:"Rating"`
	Comment    string     `json:"Comment"`
	CreateAt   time.Time  `json:"created_at"`
	UpdateAt   *time.Time `json:"updated_at"`
}

func (r *ReviewResponse) FromDomain(d domain.Review, bookTitle, readerName string) ReviewResponse {
	return ReviewResponse{
		ID:         d.ID,
		BookID:     d.BookID,
		BookTitle:  bookTitle,
		ReaderID:   d.ReaderID,
		ReaderName: readerName,
		Rating:     d.Rating,
		Comment:    d.Comment,
		CreateAt:   d.CreatedAt,
		UpdateAt:   d.UpdatedAt,
	}
}

type ReviewListResponse struct {
	Reviews       []ReviewResponse      `json:"reviews"`
	AverageRating float64               `json:"average_rating"`
	TotalRewiews  int                   `json:"total_rewiews"`
	Pagination    pagination.Pagination `json:"pagination"`
}
