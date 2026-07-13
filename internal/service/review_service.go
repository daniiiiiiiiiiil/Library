package service

import (
	"context"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"

	"github.com/jackc/pgx/v5"
)

type ReviewService struct {
	reviewRepo repository.ReviewRepository
	bookRepo   repository.BookRepository
	readerRepo repository.ReaderRepository
	txRepo     repository.TransactionRepository
}

func NewReview(
	reviewRepo repository.ReviewRepository,
	bookRepo repository.BookRepository,
	readerRepo repository.ReaderRepository,
	txRepo repository.TransactionRepository) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		bookRepo:   bookRepo,
		readerRepo: readerRepo,
		txRepo:     txRepo,
	}
}

func (r *ReviewService) CreateReview(ctx context.Context, conn *pgx.Conn, review domain.Review) (*domain.Review, error) {
	if err := review.Validate(); err != nil {
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации" + err.Error(),
		}
	}
	exists, err := r.bookRepo.Exists(ctx, conn, review.BookID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NotFoundError{
			Entity: "Book",
			ID:     review.BookID,
		}
	}

	reader, err := r.readerRepo.GetByID(ctx, conn, review.ReaderID)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Reader",
			ID:     review.ReaderID,
		}
	}
	if reader == nil {
		return nil, errors.NotFoundError{
			Entity: "Reader",
			ID:     review.ReaderID,
		}
	}

	hasBorrowed, err := r.txRepo.HasReaderBorrowedBook(ctx, conn, review.ReaderID, review.BookID)
	if err != nil {
		return nil, err
	}
	if !hasBorrowed {
		return nil, errors.BusinessError{
			Code:    "reader_has_not_borrowed_book",
			Message: "Читатель не брал эту книгу",
		}
	}

	existsReview, err := r.reviewRepo.Exists(ctx, conn, review.BookID, review.ReaderID)
	if err != nil {
		return nil, err
	}
	if existsReview {
		return nil, errors.BusinessError{
			Code:    "review_already_exists",
			Message: "Читатель уже оставлял отзыв на эту книгу",
		}
	}

	if err := review.Validate(); err != nil {
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	createdReview, err := r.reviewRepo.CreateReview(ctx, conn, review)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "create_review_error",
			Message: "Не удалось создать отзыв: " + err.Error(),
		}
	}

	avgRating, err := r.reviewRepo.GetAverageRating(ctx, conn, review.BookID)
	if err != nil {
		return nil, err
	}

	if err := r.bookRepo.UpdateRating(ctx, conn, review.BookID, avgRating); err != nil {
		return nil, err
	}

	return createdReview, nil
}

func (r *ReviewService) GetReview(ctx context.Context, conn *pgx.Conn, reviewID int) (*domain.Review, error) {
	if reviewID < 0 {
		return nil, errors.ValidationError{
			Field:   "ReviewID",
			Message: "ID отзыв меньше 0",
		}
	}
	reviewGet, err := r.reviewRepo.GetByID(ctx, conn, reviewID)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Review",
			ID:     reviewID,
		}
	}
	return &reviewGet, nil
}

func (r *ReviewService) GetReviewsByBook(ctx context.Context, conn *pgx.Conn, bookID int, limit, offset int) ([]domain.Review, int, error) {
	limitOffset(limit, offset)
	exists, err := r.bookRepo.Exists(ctx, conn, bookID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, errors.NotFoundError{
			Entity: "Book",
			ID:     bookID,
		}
	}
	getReview, err := r.reviewRepo.GetReviewsByBookID(ctx, conn, bookID, limit, offset)
	if err != nil {
		return nil, 0, errors.NotFoundError{
			Entity: "Review",
			ID:     bookID,
		}
	}
	return getReview, len(getReview), nil
}

func (r *ReviewService) GetReviewsByReader(ctx context.Context, conn *pgx.Conn, readerID int, limit, offset int) ([]domain.Review, int, error) {
	limitOffset(limit, offset)
	exists, err := r.readerRepo.Exists(ctx, conn, readerID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}
	getReaderRebiew, err := r.reviewRepo.GetReviewsByReaderID(ctx, conn, readerID, limit, offset)
	if err != nil {
		return nil, 0, errors.NotFoundError{
			Entity: "ReviewByReader",
			ID:     readerID,
		}
	}
	return getReaderRebiew, len(getReaderRebiew), nil
}

func (r *ReviewService) UpdateReview(ctx context.Context, conn *pgx.Conn, reviewID, readerID int, rating float64, comment string) (*domain.Review, error) {
	existingReview, err := r.reviewRepo.GetByID(ctx, conn, reviewID)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Review",
			ID:     reviewID,
		}
	}

	if existingReview.ReaderID != readerID {
		return nil, errors.BusinessError{
			Code:    "ErrNotAuthor",
			Message: "Только автор может редактировать свой отзыв",
		}
	}

	if rating >= 1 && rating <= 10 {
		existingReview.Rating = rating
	}
	if comment != "" {
		existingReview.Comment = comment
	}

	if err := existingReview.Validate(); err != nil {
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	if err := r.reviewRepo.Update(ctx, conn, existingReview); err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrUpdateReview",
			Message: "Не удалось обновить отзыв: " + err.Error(),
		}
	}

	avgRating, err := r.reviewRepo.GetAverageRating(ctx, conn, existingReview.BookID)
	if err != nil {
		return nil, err
	}

	if err := r.bookRepo.UpdateRating(ctx, conn, existingReview.BookID, avgRating); err != nil {
		return nil, err
	}

	return &existingReview, nil
}

func (r *ReviewService) DeleteReview(ctx context.Context, conn *pgx.Conn, reviewID int, readerID int) error {
	var review *domain.Review
	existingReview, err := r.reviewRepo.GetByID(ctx, conn, reviewID)
	if err != nil {
		return errors.NotFoundError{
			Entity: "Review",
			ID:     reviewID,
		}
	}

	if existingReview.ReaderID != readerID {
		return errors.BusinessError{
			Code:    "ErrNotAuthor",
			Message: "Только автор может редактировать свой отзыв",
		}
	}
	if err := r.reviewRepo.Delete(ctx, conn, reviewID); err != nil {
		return errors.BusinessError{
			Code:    "ErrDeleteReview",
			Message: "Не удалось удалить отзыв" + err.Error(),
		}
	}
	avgRating, err := r.reviewRepo.GetAverageRating(ctx, conn, review.BookID)
	if err != nil {
		return err
	}
	if err := r.bookRepo.UpdateRating(ctx, conn, existingReview.BookID, avgRating); err != nil {
		return errors.BusinessError{
			Code:    "ErrUpdateReview",
			Message: "Не удалось пересчитать средний рейтинг: " + err.Error(),
		}
	}
	return nil
}

func (r *ReviewService) GetBookRating(ctx context.Context, conn *pgx.Conn, bookID int) (float64, int, error) {
	exists, err := r.bookRepo.Exists(ctx, conn, bookID)
	if err != nil {
		return 0, 0, err
	}
	if !exists {
		return 0, 0, errors.NotFoundError{
			Entity: "Book",
			ID:     bookID,
		}
	}
	avg, err := r.reviewRepo.GetAverageRating(ctx, conn, bookID)
	if err != nil {
		return 0, 0, errors.NotFoundError{
			Entity: "AVGRating",
			ID:     bookID,
		}
	}
	count, err := r.reviewRepo.GetReviewCount(ctx, conn, bookID)
	if err != nil {
		return 0, 0, errors.NotFoundError{
			Entity: "CountReview",
			ID:     bookID,
		}
	}
	return avg, count, err
}

func (r *ReviewService) UpdateBookRating(ctx context.Context, conn *pgx.Conn, bookID int) error {
	avg, err := r.reviewRepo.GetAverageRating(ctx, conn, bookID)
	if err != nil {
		return errors.NotFoundError{
			Entity: "ErrReviewAVGRAting",
			ID:     bookID,
		}
	}
	if err := r.bookRepo.UpdateRating(ctx, conn, bookID, avg); err != nil {
		return errors.BusinessError{
			Code:    "ErrUpdateReview",
			Message: "Не удалось обновить рейтинг: " + err.Error(),
		}
	}
	if err := r.bookRepo.UpdateRatingAndCount(ctx, conn, bookID, 1); err != nil {
		return errors.BusinessError{
			Code: "ErrUpdateReview",
		}
	}
	return nil
}
