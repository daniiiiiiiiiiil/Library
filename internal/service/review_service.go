package service

import (
	"context"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ReviewService struct {
	reviewRepo repository.ReviewRepository
	bookRepo   repository.BookRepository
	readerRepo repository.ReaderRepository
	txRepo     repository.TransactionRepository
	logger     *zap.Logger
}

func NewReview(
	reviewRepo repository.ReviewRepository,
	bookRepo repository.BookRepository,
	readerRepo repository.ReaderRepository,
	txRepo repository.TransactionRepository,
	logger *zap.Logger,
) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		bookRepo:   bookRepo,
		readerRepo: readerRepo,
		txRepo:     txRepo,
		logger:     logger,
	}
}

func (r *ReviewService) CreateReview(ctx context.Context, conn *pgx.Conn, review domain.Review) (*domain.Review, error) {
	r.logger.Info("create review started", zap.Int("book_id", review.BookID), zap.Int("reader_id", review.ReaderID), zap.Float64("rating", review.Rating))

	if err := review.Validate(); err != nil {
		r.logger.Warn("review validation failed", zap.Int("book_id", review.BookID), zap.Int("reader_id", review.ReaderID), zap.Error(err))
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации" + err.Error(),
		}
	}

	exists, err := r.bookRepo.Exists(ctx, conn, review.BookID)
	if err != nil {
		r.logger.Error("failed to check book existence", zap.Int("book_id", review.BookID), zap.Error(err))
		return nil, err
	}
	if !exists {
		r.logger.Warn("book not found for review", zap.Int("book_id", review.BookID))
		return nil, errors.NotFoundError{
			Entity: "Book",
			ID:     review.BookID,
		}
	}

	reader, err := r.readerRepo.GetByID(ctx, conn, review.ReaderID)
	if err != nil {
		r.logger.Error("failed to get reader", zap.Int("reader_id", review.ReaderID), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Reader",
			ID:     review.ReaderID,
		}
	}
	if reader == nil {
		r.logger.Warn("reader not found for review", zap.Int("reader_id", review.ReaderID))
		return nil, errors.NotFoundError{
			Entity: "Reader",
			ID:     review.ReaderID,
		}
	}

	hasBorrowed, err := r.txRepo.HasReaderBorrowedBook(ctx, conn, review.ReaderID, review.BookID)
	if err != nil {
		r.logger.Error("failed to check if reader borrowed book", zap.Int("reader_id", review.ReaderID), zap.Int("book_id", review.BookID), zap.Error(err))
		return nil, err
	}
	if !hasBorrowed {
		r.logger.Warn("reader has not borrowed this book", zap.Int("reader_id", review.ReaderID), zap.Int("book_id", review.BookID))
		return nil, errors.BusinessError{
			Code:    "reader_has_not_borrowed_book",
			Message: "Читатель не брал эту книгу",
		}
	}

	existsReview, err := r.reviewRepo.Exists(ctx, conn, review.BookID, review.ReaderID)
	if err != nil {
		r.logger.Error("failed to check review existence", zap.Int("book_id", review.BookID), zap.Int("reader_id", review.ReaderID), zap.Error(err))
		return nil, err
	}
	if existsReview {
		r.logger.Warn("review already exists", zap.Int("book_id", review.BookID), zap.Int("reader_id", review.ReaderID))
		return nil, errors.BusinessError{
			Code:    "review_already_exists",
			Message: "Читатель уже оставлял отзыв на эту книгу",
		}
	}

	if err := review.Validate(); err != nil {
		r.logger.Warn("review validation failed", zap.Int("book_id", review.BookID), zap.Error(err))
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	createdReview, err := r.reviewRepo.CreateReview(ctx, conn, review)
	if err != nil {
		r.logger.Error("failed to create review", zap.Int("book_id", review.BookID), zap.Int("reader_id", review.ReaderID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "create_review_error",
			Message: "Не удалось создать отзыв: " + err.Error(),
		}
	}

	avgRating, err := r.reviewRepo.GetAverageRating(ctx, conn, review.BookID)
	if err != nil {
		r.logger.Error("failed to get average rating", zap.Int("book_id", review.BookID), zap.Error(err))
		return nil, err
	}

	if err := r.bookRepo.UpdateRating(ctx, conn, review.BookID, avgRating); err != nil {
		r.logger.Error("failed to update book rating", zap.Int("book_id", review.BookID), zap.Float64("avg_rating", avgRating), zap.Error(err))
		return nil, err
	}

	r.logger.Info("review created successfully", zap.Int("review_id", createdReview.ID), zap.Int("book_id", createdReview.BookID), zap.Int("reader_id", createdReview.ReaderID), zap.Float64("rating", createdReview.Rating))
	return createdReview, nil
}

func (r *ReviewService) GetReview(ctx context.Context, conn *pgx.Conn, reviewID int) (*domain.Review, error) {
	r.logger.Debug("get review started", zap.Int("review_id", reviewID))

	if reviewID < 0 {
		r.logger.Warn("invalid review id", zap.Int("review_id", reviewID))
		return nil, errors.ValidationError{
			Field:   "ReviewID",
			Message: "ID отзыв меньше 0",
		}
	}

	reviewGet, err := r.reviewRepo.GetByID(ctx, conn, reviewID)
	if err != nil {
		r.logger.Warn("review not found", zap.Int("review_id", reviewID), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Review",
			ID:     reviewID,
		}
	}

	r.logger.Debug("get review finished", zap.Int("review_id", reviewGet.ID))
	return &reviewGet, nil
}

func (r *ReviewService) GetReviewsByBook(ctx context.Context, conn *pgx.Conn, bookID int, limit, offset int) ([]domain.Review, int, error) {
	limit, offset = limitOffset(limit, offset)
	r.logger.Debug("get reviews by book started", zap.Int("book_id", bookID), zap.Int("limit", limit), zap.Int("offset", offset))

	exists, err := r.bookRepo.Exists(ctx, conn, bookID)
	if err != nil {
		r.logger.Error("failed to check book existence", zap.Int("book_id", bookID), zap.Error(err))
		return nil, 0, err
	}
	if !exists {
		r.logger.Warn("book not found for get reviews", zap.Int("book_id", bookID))
		return nil, 0, errors.NotFoundError{
			Entity: "Book",
			ID:     bookID,
		}
	}

	getReview, err := r.reviewRepo.GetReviewsByBookID(ctx, conn, bookID, limit, offset)
	if err != nil {
		r.logger.Error("failed to get reviews by book", zap.Int("book_id", bookID), zap.Error(err))
		return nil, 0, errors.NotFoundError{
			Entity: "Review",
			ID:     bookID,
		}
	}

	r.logger.Debug("get reviews by book finished", zap.Int("book_id", bookID), zap.Int("returned", len(getReview)))
	return getReview, len(getReview), nil
}

func (r *ReviewService) GetReviewsByReader(ctx context.Context, conn *pgx.Conn, readerID int, limit, offset int) ([]domain.Review, int, error) {
	limit, offset = limitOffset(limit, offset)
	r.logger.Debug("get reviews by reader started", zap.Int("reader_id", readerID), zap.Int("limit", limit), zap.Int("offset", offset))

	exists, err := r.readerRepo.Exists(ctx, conn, readerID)
	if err != nil {
		r.logger.Error("failed to check reader existence", zap.Int("reader_id", readerID), zap.Error(err))
		return nil, 0, err
	}
	if !exists {
		r.logger.Warn("reader not found for get reviews", zap.Int("reader_id", readerID))
		return nil, 0, errors.NotFoundError{
			Entity: "Reader",
			ID:     readerID,
		}
	}

	getReaderReview, err := r.reviewRepo.GetReviewsByReaderID(ctx, conn, readerID, limit, offset)
	if err != nil {
		r.logger.Error("failed to get reviews by reader", zap.Int("reader_id", readerID), zap.Error(err))
		return nil, 0, errors.NotFoundError{
			Entity: "ReviewByReader",
			ID:     readerID,
		}
	}

	r.logger.Debug("get reviews by reader finished", zap.Int("reader_id", readerID), zap.Int("returned", len(getReaderReview)))
	return getReaderReview, len(getReaderReview), nil
}

func (r *ReviewService) UpdateReview(ctx context.Context, conn *pgx.Conn, reviewID, readerID int, rating float64, comment string) (*domain.Review, error) {
	r.logger.Info("update review started", zap.Int("review_id", reviewID), zap.Int("reader_id", readerID), zap.Float64("rating", rating))

	existingReview, err := r.reviewRepo.GetByID(ctx, conn, reviewID)
	if err != nil {
		r.logger.Warn("review not found for update", zap.Int("review_id", reviewID), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Review",
			ID:     reviewID,
		}
	}

	if existingReview.ReaderID != readerID {
		r.logger.Warn("user is not author of review", zap.Int("review_id", reviewID), zap.Int("reader_id", readerID), zap.Int("author_id", existingReview.ReaderID))
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
		r.logger.Warn("review validation failed on update", zap.Int("review_id", reviewID), zap.Error(err))
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	if err := r.reviewRepo.Update(ctx, conn, existingReview); err != nil {
		r.logger.Error("failed to update review", zap.Int("review_id", reviewID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrUpdateReview",
			Message: "Не удалось обновить отзыв: " + err.Error(),
		}
	}

	avgRating, err := r.reviewRepo.GetAverageRating(ctx, conn, existingReview.BookID)
	if err != nil {
		r.logger.Error("failed to get average rating after update", zap.Int("book_id", existingReview.BookID), zap.Error(err))
		return nil, err
	}

	if err := r.bookRepo.UpdateRating(ctx, conn, existingReview.BookID, avgRating); err != nil {
		r.logger.Error("failed to update book rating after review update", zap.Int("book_id", existingReview.BookID), zap.Float64("avg_rating", avgRating), zap.Error(err))
		return nil, err
	}

	r.logger.Info("review updated successfully", zap.Int("review_id", reviewID))
	return &existingReview, nil
}

func (r *ReviewService) DeleteReview(ctx context.Context, conn *pgx.Conn, reviewID int, readerID int) error {
	r.logger.Info("delete review started", zap.Int("review_id", reviewID), zap.Int("reader_id", readerID))

	existingReview, err := r.reviewRepo.GetByID(ctx, conn, reviewID)
	if err != nil {
		r.logger.Warn("review not found for delete", zap.Int("review_id", reviewID), zap.Error(err))
		return errors.NotFoundError{
			Entity: "Review",
			ID:     reviewID,
		}
	}

	if existingReview.ReaderID != readerID {
		r.logger.Warn("user is not author of review", zap.Int("review_id", reviewID), zap.Int("reader_id", readerID), zap.Int("author_id", existingReview.ReaderID))
		return errors.BusinessError{
			Code:    "ErrNotAuthor",
			Message: "Только автор может редактировать свой отзыв",
		}
	}

	if err := r.reviewRepo.Delete(ctx, conn, reviewID); err != nil {
		r.logger.Error("failed to delete review", zap.Int("review_id", reviewID), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrDeleteReview",
			Message: "Не удалось удалить отзыв" + err.Error(),
		}
	}

	avgRating, err := r.reviewRepo.GetAverageRating(ctx, conn, existingReview.BookID)
	if err != nil {
		r.logger.Error("failed to get average rating after delete", zap.Int("book_id", existingReview.BookID), zap.Error(err))
		return err
	}

	if err := r.bookRepo.UpdateRating(ctx, conn, existingReview.BookID, avgRating); err != nil {
		r.logger.Error("failed to update book rating after review delete", zap.Int("book_id", existingReview.BookID), zap.Float64("avg_rating", avgRating), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrUpdateReview",
			Message: "Не удалось пересчитать средний рейтинг: " + err.Error(),
		}
	}

	r.logger.Info("review deleted successfully", zap.Int("review_id", reviewID))
	return nil
}

func (r *ReviewService) GetBookRating(ctx context.Context, conn *pgx.Conn, bookID int) (float64, int, error) {
	r.logger.Debug("get book rating started", zap.Int("book_id", bookID))

	exists, err := r.bookRepo.Exists(ctx, conn, bookID)
	if err != nil {
		r.logger.Error("failed to check book existence", zap.Int("book_id", bookID), zap.Error(err))
		return 0, 0, err
	}
	if !exists {
		r.logger.Warn("book not found for get rating", zap.Int("book_id", bookID))
		return 0, 0, errors.NotFoundError{
			Entity: "Book",
			ID:     bookID,
		}
	}

	avg, err := r.reviewRepo.GetAverageRating(ctx, conn, bookID)
	if err != nil {
		r.logger.Error("failed to get average rating", zap.Int("book_id", bookID), zap.Error(err))
		return 0, 0, errors.NotFoundError{
			Entity: "AVGRating",
			ID:     bookID,
		}
	}

	count, err := r.reviewRepo.GetReviewCount(ctx, conn, bookID)
	if err != nil {
		r.logger.Error("failed to get review count", zap.Int("book_id", bookID), zap.Error(err))
		return 0, 0, errors.NotFoundError{
			Entity: "CountReview",
			ID:     bookID,
		}
	}

	r.logger.Debug("get book rating finished", zap.Int("book_id", bookID), zap.Float64("avg_rating", avg), zap.Int("review_count", count))
	return avg, count, err
}

func (r *ReviewService) UpdateBookRating(ctx context.Context, conn *pgx.Conn, bookID int) error {
	r.logger.Info("update book rating started", zap.Int("book_id", bookID))

	avg, err := r.reviewRepo.GetAverageRating(ctx, conn, bookID)
	if err != nil {
		r.logger.Error("failed to get average rating", zap.Int("book_id", bookID), zap.Error(err))
		return errors.NotFoundError{
			Entity: "ErrReviewAVGRAting",
			ID:     bookID,
		}
	}

	if err := r.bookRepo.UpdateRating(ctx, conn, bookID, avg); err != nil {
		r.logger.Error("failed to update book rating", zap.Int("book_id", bookID), zap.Float64("avg_rating", avg), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrUpdateReview",
			Message: "Не удалось обновить рейтинг: " + err.Error(),
		}
	}

	if err := r.bookRepo.UpdateRatingAndCount(ctx, conn, bookID, 1); err != nil {
		r.logger.Error("failed to update rating and count", zap.Int("book_id", bookID), zap.Error(err))
		return errors.BusinessError{
			Code: "ErrUpdateReview",
		}
	}

	r.logger.Info("book rating updated successfully", zap.Int("book_id", bookID), zap.Float64("avg_rating", avg))
	return nil
}
