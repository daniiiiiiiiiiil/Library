package postgres

import (
	"context"
	"fmt"
	"library/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

type ReviewRepository struct{}

func (r *ReviewRepository) CreateReview(ctx context.Context, conn *pgx.Conn, review domain.Review) (*domain.Review, error) {
	sqlQuery := `
        INSERT INTO reviews (book_id, reader_id, rating, comment, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING review_id
    `
	var id int
	err := conn.QueryRow(ctx, sqlQuery,
		review.BookID,
		review.ReaderID,
		review.Rating,
		review.Comment,
		time.Now(),
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	review.ID = id
	return &review, nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Review, error) {
	sqlQuery := `
        SELECT review_id, book_id, reader_id, rating, comment, created_at, updated_at
        FROM reviews
        WHERE review_id = $1
    `
	var review domain.Review
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&review.ID,
		&review.BookID,
		&review.ReaderID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Review{}, fmt.Errorf("review с ID %d нету", id)
		}
		return domain.Review{}, err
	}
	return review, nil
}

func (r *ReviewRepository) Update(ctx context.Context, conn *pgx.Conn, review domain.Review) error {
	sqlQuery := `
		UPDATE reviews
		SET rating = $1, comment = $2, updated_at = $3
		WHERE review_id = $4
	`
	_, err := conn.Exec(ctx, sqlQuery,
		review.Rating,
		review.Comment,
		time.Now(),
		review.ID,
	)
	return err
}

func (r *ReviewRepository) Delete(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
		DELETE FROM reviews
		WHERE review_id = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, id)
	return err
}

func (r *ReviewRepository) GetReviewsByBookID(ctx context.Context, conn *pgx.Conn, bookID int, limit, offset int) ([]domain.Review, error) {
	sqlQuery := `
		SELECT review_id, book_id, reader_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE book_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := conn.Query(ctx, sqlQuery, bookID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var review domain.Review
		if err := rows.Scan(
			&review.ID,
			&review.BookID,
			&review.ReaderID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (r *ReviewRepository) GetReviewsByReaderID(ctx context.Context, conn *pgx.Conn, readerID int, limit, offset int) ([]domain.Review, error) {
	sqlQuery := `
		SELECT review_id, book_id, reader_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE reader_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := conn.Query(ctx, sqlQuery, readerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var review domain.Review
		if err := rows.Scan(
			&review.ID,
			&review.BookID,
			&review.ReaderID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (r *ReviewRepository) GetAverageRating(ctx context.Context, conn *pgx.Conn, bookID int) (float64, error) {
	sqlQuery := `
		SELECT COALESCE(ROUND(AVG(rating)::NUMERIC, 2), 0)
		FROM reviews
		WHERE book_id = $1
	`
	var avgRating float64
	err := conn.QueryRow(ctx, sqlQuery, bookID).Scan(&avgRating)
	if err != nil {
		return 0, err
	}
	return avgRating, nil
}

func (r *ReviewRepository) GetReviewCount(ctx context.Context, conn *pgx.Conn, bookID int) (int, error) {
	sqlQuery := `
		SELECT COUNT(*)
		FROM reviews
		WHERE book_id = $1
	`
	var count int
	err := conn.QueryRow(ctx, sqlQuery, bookID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ReviewRepository) Exists(ctx context.Context, conn *pgx.Conn, bookID, readerID int) (bool, error) {
	sqlQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM reviews 
			WHERE book_id = $1 AND reader_id = $2
		)
	`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, bookID, readerID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *ReviewRepository) List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Review, error) {
	sqlQuery := `
		SELECT review_id, book_id, reader_id, rating, comment, created_at, updated_at
		FROM reviews
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var review domain.Review
		if err := rows.Scan(
			&review.ID,
			&review.BookID,
			&review.ReaderID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (r *ReviewRepository) Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Review, int, error) {
	allowedColumns := map[string]bool{
		"comment": true,
		"rating":  true,
	}
	if !allowedColumns[column] {
		return nil, 0, fmt.Errorf("недопустимая колонка: %s", column)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT review_id, book_id, reader_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE %s::text ILIKE '%%' || $1 || '%%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, column)

	rows, err := conn.Query(ctx, sqlQuery, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []domain.Review
	count := 0
	for rows.Next() {
		var review domain.Review
		if err := rows.Scan(
			&review.ID,
			&review.BookID,
			&review.ReaderID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		reviews = append(reviews, review)
		count++
	}
	return reviews, count, nil
}
