package postgres

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/repository"
	"time"

	"github.com/jackc/pgx/v5"
)

// Проверяем, что BookRepository реализует интерфейс
var _ repository.BookRepository = (*BookRepository)(nil)

type BookRepository struct{}

func (r *BookRepository) CreateBook(ctx context.Context, conn *pgx.Conn, book domain.Book) (*domain.Book, error) {
	sqlQuery := `
	INSERT INTO books(title, isbn, year, publisher_id, description, cover_image, avg_rating, reviews_count, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING book_id
`
	var id int
	err := conn.QueryRow(ctx, sqlQuery,
		book.Title,
		book.ISBN,
		book.Year,
		book.PublisherID,
		book.Description,
		book.CoverImageURL,
		book.AvgRating,
		book.ReviewsCount,
		time.Now(),
		time.Now(),
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	book.ID = id
	return &book, nil
}

func (r *BookRepository) GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Book, error) {
	sqlQuery := `
	SELECT book_id, title, isbn, year, publisher_id, description, cover_image, avg_rating, reviews_count, created_at, updated_at
	FROM books
	WHERE book_id = $1
`
	var book domain.Book
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&book.ID,
		&book.Title,
		&book.ISBN,
		&book.Year,
		&book.PublisherID,
		&book.Description,
		&book.CoverImageURL,
		&book.AvgRating,
		&book.ReviewsCount,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if err != nil {
		return domain.Book{}, err
	}
	return book, nil
}

func (r *BookRepository) GetByISBN(ctx context.Context, conn *pgx.Conn, isbn string) (domain.Book, error) {
	sqlQuery := `
		SELECT book_id, title, isbn, year, publisher_id, description, cover_image, avg_rating, reviews_count, created_at, updated_at
		FROM books
		WHERE isbn = $1
	`
	var book domain.Book
	err := conn.QueryRow(ctx, sqlQuery, isbn).Scan(
		&book.ID,
		&book.Title,
		&book.ISBN,
		&book.Year,
		&book.PublisherID,
		&book.Description,
		&book.CoverImageURL,
		&book.AvgRating,
		&book.ReviewsCount,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if err != nil {
		return domain.Book{}, err
	}
	return book, nil
}

func (r *BookRepository) Update(ctx context.Context, conn *pgx.Conn, id int, book domain.Book) error {
	sqlQuery := `
		UPDATE books
		SET title = $1, isbn = $2, year = $3, publisher_id = $4, description = $5, 
		    cover_image = $6, avg_rating = $7, reviews_count = $8, updated_at = $9
		WHERE book_id = $10
	`
	_, err := conn.Exec(ctx, sqlQuery,
		book.Title, book.ISBN, book.Year, book.PublisherID,
		book.Description, book.CoverImageURL, book.AvgRating, book.ReviewsCount,
		time.Now(), id,
	)
	return err
}

func (r *BookRepository) Delete(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
		DELETE FROM books
		WHERE book_id = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, id)
	return err
}

func (r *BookRepository) List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, error) {
	sqlQuery := `
		SELECT book_id, title, isbn, year, publisher_id, description, cover_image, 
		       avg_rating, reviews_count, created_at, updated_at
		FROM books
		ORDER BY title ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]domain.Book, 0)
	for rows.Next() {
		var book domain.Book
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.ISBN,
			&book.Year,
			&book.PublisherID,
			&book.Description,
			&book.CoverImageURL,
			&book.AvgRating,
			&book.ReviewsCount,
			&book.CreatedAt,
			&book.UpdatedAt,
		); err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	return books, nil
}

func (r *BookRepository) Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Book, int, error) {
	allowedColumns := map[string]bool{
		"title": true, "isbn": true, "year": true, "description": true,
	}
	if !allowedColumns[column] {
		return nil, 0, fmt.Errorf("недопустимая колонка: %s", column)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT book_id, title, isbn, year, publisher_id, description, cover_image, 
		       avg_rating, reviews_count, created_at, updated_at
		FROM books
		WHERE %s ILIKE '%%' || $1 || '%%'
		ORDER BY title ASC
		LIMIT $2 OFFSET $3
	`, column)

	rows, err := conn.Query(ctx, sqlQuery, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	books := make([]domain.Book, 0)
	count := 0
	for rows.Next() {
		var book domain.Book
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.ISBN,
			&book.Year,
			&book.PublisherID,
			&book.Description,
			&book.CoverImageURL,
			&book.AvgRating,
			&book.ReviewsCount,
			&book.CreatedAt,
			&book.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		books = append(books, book)
		count++
	}
	return books, count, nil
}

func (r *BookRepository) GetPopular(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, error) {
	sqlQuery := `
		SELECT 
			b.book_id, b.title, b.isbn, b.year, b.publisher_id, b.description, 
			b.cover_image, b.avg_rating, b.reviews_count, b.created_at, b.updated_at,
			COUNT(t.transaction_id) AS borrow_count
		FROM books b
		LEFT JOIN book_copies bc ON b.book_id = bc.book_id
		LEFT JOIN transactions t ON bc.book_copy_id = t.copy_id AND t.status = 'active'
		GROUP BY b.book_id, b.title, b.isbn, b.year, b.publisher_id, 
		         b.description, b.cover_image, b.avg_rating, 
		         b.reviews_count, b.created_at, b.updated_at
		ORDER BY borrow_count DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []domain.Book
	for rows.Next() {
		var book domain.Book
		var borrowCount int
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.ISBN,
			&book.Year,
			&book.PublisherID,
			&book.Description,
			&book.CoverImageURL,
			&book.AvgRating,
			&book.ReviewsCount,
			&book.CreatedAt,
			&book.UpdatedAt,
			&borrowCount,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	return books, nil
}

func (r *BookRepository) Exists(ctx context.Context, conn *pgx.Conn, bookID int) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM books WHERE book_id = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, bookID).Scan(&exists)
	return exists, err
}

func (r *BookRepository) ExistsByISBN(ctx context.Context, conn *pgx.Conn, isbn string) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM books WHERE isbn = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, isbn).Scan(&exists)
	return exists, err
}

func (r *BookRepository) Count(ctx context.Context, conn *pgx.Conn) (int, error) {
	sqlQuery := `SELECT COUNT(*) FROM books`
	var count int
	err := conn.QueryRow(ctx, sqlQuery).Scan(&count)
	return count, err
}

func (r *BookRepository) GetByPublisherID(ctx context.Context, conn *pgx.Conn, publisherID, limit, offset int) ([]domain.Book, error) {
	sqlQuery := `
		SELECT book_id, title, isbn, year, publisher_id, description, cover_image, 
		       avg_rating, reviews_count, created_at, updated_at
		FROM books
		WHERE publisher_id = $1
		ORDER BY title ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := conn.Query(ctx, sqlQuery, publisherID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []domain.Book
	for rows.Next() {
		var book domain.Book
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.ISBN,
			&book.Year,
			&book.PublisherID,
			&book.Description,
			&book.CoverImageURL,
			&book.AvgRating,
			&book.ReviewsCount,
			&book.CreatedAt,
			&book.UpdatedAt,
		); err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	return books, nil
}

func (r *BookRepository) CountByPublisherID(ctx context.Context, conn *pgx.Conn, publisherID int) (int, error) {
	sqlQuery := `
		SELECT COUNT(*)
		FROM books
		WHERE publisher_id = $1
	`
	var count int
	err := conn.QueryRow(ctx, sqlQuery, publisherID).Scan(&count)
	return count, err
}

func (r *BookRepository) UpdateRating(ctx context.Context, conn *pgx.Conn, bookID int, avgRating float64) error {
	sqlQuery := `
		UPDATE books
		SET avg_rating = $1
		WHERE book_id = $2
	`
	_, err := conn.Exec(ctx, sqlQuery, avgRating, bookID)
	return err
}

func (r *BookRepository) UpdateRatingAndCount(ctx context.Context, conn *pgx.Conn, bookID int, reviewsCount int) error {
	sqlQuery := `
		UPDATE books
		SET reviews_count = $1
		WHERE book_id = $2
	`
	_, err := conn.Exec(ctx, sqlQuery, reviewsCount, bookID)
	return err
}
