package postgres

import (
	"context"
	"fmt"
	"library/internal/domain"

	"github.com/jackc/pgx/v5"
)

func CreateAuthor(ctx context.Context, conn *pgx.Conn, author *domain.Author) (*domain.Author, error) {
	sqlQuery := `
		INSERT INTO authors (first_name, last_name, biography, birth_date)
		VALUES ($1, $2, $3, $4)
		RETURNING authors_id
	`
	var id int
	err := conn.QueryRow(ctx, sqlQuery,
		author.First_name,
		author.Last_name,
		author.Biography,
		author.Birthday,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	author.ID = id
	return author, nil
}

func GetByIDAuthor(ctx context.Context, conn *pgx.Conn, id int) (domain.Author, error) {
	sqlQuery := `
		SELECT authors_id, first_name, last_name, biography, birth_date
		FROM authors
		WHERE authors_id = $1
	`
	var author domain.Author
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&author.ID,
		&author.First_name,
		&author.Last_name,
		&author.Biography,
		&author.Birthday,
	)
	if err != nil {
		return domain.Author{}, err
	}
	return author, nil
}

func UpdateAuthor(ctx context.Context, conn *pgx.Conn, author domain.Author) error {
	sqlQuery := `
		UPDATE authors
		SET first_name = $1, last_name = $2, biography = $3, birth_date = $4
		WHERE authors_id = $5
	`
	_, err := conn.Exec(ctx, sqlQuery,
		author.First_name,
		author.Last_name,
		author.Biography,
		author.Birthday,
		author.ID,
	)
	return err
}

func DeleteAuthor(ctx context.Context, conn *pgx.Conn, id int) error {
	sqlQuery := `
		DELETE FROM authors
		WHERE authors_id = $1
	`
	_, err := conn.Exec(ctx, sqlQuery, id)
	return err
}

func ListAuthor(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Author, error) {
	sqlQuery := `
		SELECT authors_id, first_name, last_name, biography, birth_date
		FROM authors
		ORDER BY last_name ASC, first_name ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []domain.Author
	for rows.Next() {
		var author domain.Author
		if err := rows.Scan(
			&author.ID,
			&author.First_name,
			&author.Last_name,
			&author.Biography,
			&author.Birthday,
		); err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}
	return authors, nil
}

func GetByBookIDAuthor(ctx context.Context, conn *pgx.Conn, bookID int) ([]domain.Author, error) {
	sqlQuery := `
		SELECT a.authors_id, a.first_name, a.last_name, a.biography, a.birth_date
		FROM authors a
		JOIN book_authors ba ON a.authors_id = ba.authors_id
		WHERE ba.book_id = $1
		ORDER BY a.last_name ASC, a.first_name ASC
	`
	rows, err := conn.Query(ctx, sqlQuery, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []domain.Author
	for rows.Next() {
		var author domain.Author
		if err := rows.Scan(
			&author.ID,
			&author.First_name,
			&author.Last_name,
			&author.Biography,
			&author.Birthday,
		); err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}
	return authors, nil
}

func SearchAuthor(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Author, int, error) {
	allowedColumns := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"biography":  true,
	}
	if !allowedColumns[column] {
		return nil, 0, fmt.Errorf("недопустимая колонка: %s", column)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT authors_id, first_name, last_name, biography, birth_date
		FROM authors
		WHERE %s ILIKE '%%' || $1 || '%%'
		ORDER BY last_name ASC, first_name ASC
		LIMIT $2 OFFSET $3
	`, column)

	rows, err := conn.Query(ctx, sqlQuery, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var authors []domain.Author
	count := 0
	for rows.Next() {
		var author domain.Author
		if err := rows.Scan(
			&author.ID,
			&author.First_name,
			&author.Last_name,
			&author.Biography,
			&author.Birthday,
		); err != nil {
			return nil, 0, err
		}
		authors = append(authors, author)
		count++
	}
	return authors, count, nil
}

func ExistsAuthor(ctx context.Context, conn *pgx.Conn, authorID int) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM authors WHERE authors_id = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, authorID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CreateBookAuthor(ctx context.Context, conn *pgx.Conn, bookID, authorID int) error {
	sqlQuery := `
        INSERT INTO book_authors (book_id, author_id)
        VALUES ($1, $2)
    `
	_, err := conn.Exec(ctx, sqlQuery, bookID, authorID)
	return err
}

func DeleteBookAuthorsByBookID(ctx context.Context, conn *pgx.Conn, bookID int) error {
	sqlQuery := `
        DELETE FROM book_authors
        WHERE book_id = $1
    `
	_, err := conn.Exec(ctx, sqlQuery, bookID)
	return err
}

func ExistsByName(ctx context.Context, conn *pgx.Conn, firstName, lastName string) (bool, error) {
	sqlQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM authors 
			WHERE first_name = $1 AND last_name = $2
		)
	`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, firstName, lastName).Scan(&exists)
	return exists, err
}

func ExistsByNameExcludeID(ctx context.Context, conn *pgx.Conn, firstName, lastName string, excludeID int) (bool, error) {
	sqlQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM authors 
			WHERE first_name = $1 AND last_name = $2 AND authors_id != $3
		)
	`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, firstName, lastName, excludeID).Scan(&exists)
	return exists, err
}

func CountAuthor(ctx context.Context, conn *pgx.Conn) (int, error) {
	sqlQuery := `SELECT COUNT(*) FROM authors`
	var count int
	err := conn.QueryRow(ctx, sqlQuery).Scan(&count)
	return count, err
}

func GetBooksByAuthorID(ctx context.Context, conn *pgx.Conn, authorID int) ([]domain.Book, error) {
	sqlQuery := `
		SELECT b.book_id, b.title, b.isbn, b.year, b.publisher_id, 
		       b.description, b.cover_image, b.avg_rating, 
		       b.reviews_count, b.created_at, b.updated_at
		FROM books b
		JOIN book_authors ba ON b.book_id = ba.book_id
		WHERE ba.authors_id = $1
	`
	rows, err := conn.Query(ctx, sqlQuery, authorID)
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
