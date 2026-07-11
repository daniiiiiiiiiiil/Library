package postgres

import (
	"context"
	"fmt"
	"library/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

func CreateBook(ctx context.Context, conn *pgx.Conn, book domain.Book) error {
	sqlQuery := `
	INSERT INTO books(title,isbn,year,publisher_id,description,cover_image,avg_rating,reviews_count,created_at,updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
`
	_, err := conn.Exec(ctx, sqlQuery,
		book.Title,
		book.ISBN,
		book.Year,
		book.PublisherID,
		book.Description,
		book.CoverImageURL,
		book.AvgRating,
		book.ReviewsCount,
		time.Now(),
		book.UpdatedAt)
	return err
}

func GetByIDBook(ctx context.Context, conn *pgx.Conn, id int) (domain.Book, error) {
	sqlQuery := `
	SELECT * 
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
		&book.CoverImageURL,
		&book.AvgRating,
		&book.ReviewsCount,
		&book.CreatedAt,
		&book.UpdatedAt)
	return book, err
}

func GetByISBNBook(ctx context.Context, conn *pgx.Conn, isbn string) (domain.Book, error) {
	sqlQuery := `
		SELECT * 
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
		&book.CoverImageURL,
		&book.AvgRating,
		&book.ReviewsCount,
		&book.CreatedAt,
		&book.UpdatedAt)
	return book, err
}

func UpdateBook(ctx context.Context, conn *pgx.Conn, id int, book domain.Book) error {
	sqlQuery := `
		UPDATE books
		SET title = $1, isbn = $2, year = $3,publisher_id = $4,description = $5,cover_image = $6,avg_rating = $7,reviews_count = $8,updated_at = $9
		WHERE book_id  = $10`
	_, err := conn.Exec(ctx, sqlQuery, book.Title, book.ISBN, book.Year, book.PublisherID, book.Description, book.CoverImageURL, book.AvgRating, book.ReviewsCount, time.Now(), id)
	return err
}

func DeleteBook(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
		DELETE FROM books
		WHERE book_id = $1`
	_, err := conn.Exec(ctx, sqlQuery, id)
	return err
}

func ListBook(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, error) {
	sqlQuery := `
		SELECT *
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
			&book.CoverImageURL,
			&book.AvgRating,
			&book.ReviewsCount,
			&book.CreatedAt,
			&book.UpdatedAt); err != nil {
			return nil, err
		}
		books = append(books, book)
		printBook(book)
	}
	return books, nil
}

func printBook(book domain.Book) {
	fmt.Println("----------------------------------------------")
	fmt.Println("id:", book.ID)
	fmt.Println("title:", book.Title)
	fmt.Println("ISBN:", book.ISBN)
	fmt.Println("year:", book.Year)
	fmt.Println("publisherID:", book.PublisherID)
	fmt.Println("description:", book.Description)
	fmt.Println("coverImageURL:", book.CoverImageURL)
	fmt.Println("avgRating:", book.AvgRating)
	fmt.Println("reviews_count:", book.ReviewsCount)
	fmt.Println("createdAt:", book.CreatedAt)
	fmt.Println("updatedAt:", book.UpdatedAt)
}

func SearchBook(ctx context.Context, conn *pgx.Conn, collum, search string, limit, offset int) ([]domain.Book, int, error) {
	allowedColumns := map[string]bool{
		"title": true, "isbn": true, "year": true, "description": true,
	}
	if !allowedColumns[collum] {
		return nil, 0, fmt.Errorf("недопустимая колонка: %s", collum)
	}
	sqlQuery := fmt.Sprintf(`
		SELECT *
		FROM books
		WHERE %s LIKE '%%' || $1 || '%%'
		ORDER BY %s ASC
		LIMIT $2 OFFSET $3`, collum, collum)
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
			&book.CoverImageURL,
			&book.AvgRating,
			&book.ReviewsCount,
			&book.CreatedAt,
			&book.UpdatedAt); err != nil {
			return nil, 0, err
		}
		books = append(books, book)
		count++
		printBook(book)
		fmt.Println("Количество найденных книг", count)
	}
	return books, count, nil
}

func GetPopularBook(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, error) {
	sqlQuery := `
        SELECT 
            b.book_id,
            b.title,
            b.isbn,
            b.year,
            b.publisher_id,
            b.description,
            b.cover_image,
            b.avg_rating,
            b.reviews_count,
            b.created_at,
            b.updated_at,
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

func ExistsBook(ctx context.Context, conn *pgx.Conn, isbn string) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT * FROM books WHERE isbn = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, isbn).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
func Count(ctx context.Context, conn *pgx.Conn) (int, error) {
	sqlQuery := `SELECT COUNT(*) FROM books`
	var count int
	err := conn.QueryRow(ctx, sqlQuery).Scan(&count)
	return count, err
}
