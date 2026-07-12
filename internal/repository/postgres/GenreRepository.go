package postgres

import (
	"context"
	"fmt"
	"library/internal/domain"

	"github.com/jackc/pgx/v5"
)

func CreateGenre(ctx context.Context, conn *pgx.Conn, genre domain.Genre) (*domain.Genre, error) {
	sqlQuery := `
		INSERT INTO genres (name, parent_id)
		VALUES ($1, $2)
		RETURNING genres_id
	`
	var id int
	err := conn.QueryRow(ctx, sqlQuery, genre.Name, genre.ParentID).Scan(&id)
	if err != nil {
		return nil, err
	}

	genre.ID = id
	return &genre, nil
}

func GetByIDGenre(ctx context.Context, conn *pgx.Conn, id int) (*domain.Genre, error) {
	sqlQuery := `
		SELECT genres_id, name, parent_id
		FROM genres
		WHERE genres_id = $1
	`
	var genre domain.Genre
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(
		&genre.ID,
		&genre.Name,
		&genre.ParentID,
	)
	if err != nil {
		return &domain.Genre{}, err
	}
	return &genre, nil
}

func UpdateGenre(ctx context.Context, conn *pgx.Conn, genre domain.Genre) error {
	sqlQuery := `
		UPDATE genres
		SET name = $1, parent_id = $2
		WHERE genres_id = $3
	`
	_, err := conn.Exec(ctx, sqlQuery, genre.Name, genre.ParentID, genre.ID)
	return err
}

func DeleteGenre(ctx context.Context, conn *pgx.Conn, id int) (*domain.Genre, error) {
	var genre domain.Genre
	getQuery := `SELECT genres_id, name, parent_id FROM genres WHERE genres_id = $1`
	err := conn.QueryRow(ctx, getQuery, id).Scan(&genre.ID, &genre.Name, &genre.ParentID)
	if err != nil {
		return nil, err
	}

	deleteQuery := `DELETE FROM genres WHERE genres_id = $1`
	_, err = conn.Exec(ctx, deleteQuery, id)
	if err != nil {
		return nil, err
	}

	return &genre, nil
}

func ListGenres(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Genre, error) {
	sqlQuery := `
		SELECT genres_id, name, parent_id
		FROM genres
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := conn.Query(ctx, sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []domain.Genre
	for rows.Next() {
		var genre domain.Genre
		if err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&genre.ParentID,
		); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	return genres, nil
}

func ExistsGenre(ctx context.Context, conn *pgx.Conn, id int) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM genres WHERE genres_id = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GetSubGenres(ctx context.Context, conn *pgx.Conn, parentID int) ([]domain.Genre, error) {
	sqlQuery := `
		SELECT genres_id, name, parent_id
		FROM genres
		WHERE parent_id = $1
		ORDER BY name ASC
	`
	rows, err := conn.Query(ctx, sqlQuery, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []domain.Genre
	for rows.Next() {
		var genre domain.Genre
		if err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&genre.ParentID,
		); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	return genres, nil
}

func GetRootGenres(ctx context.Context, conn *pgx.Conn) ([]domain.Genre, error) {
	sqlQuery := `
		SELECT genres_id, name, parent_id
		FROM genres
		WHERE parent_id IS NULL
		ORDER BY name ASC
	`
	rows, err := conn.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []domain.Genre
	for rows.Next() {
		var genre domain.Genre
		if err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&genre.ParentID,
		); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	return genres, nil
}

func SearchGenre(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Genre, int, error) {
	allowedColumns := map[string]bool{
		"name": true,
	}
	if !allowedColumns[column] {
		return nil, 0, fmt.Errorf("недопустимая колонка: %s", column)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT genres_id, name, parent_id
		FROM genres
		WHERE %s ILIKE '%%' || $1 || '%%'
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`, column)

	rows, err := conn.Query(ctx, sqlQuery, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var genres []domain.Genre
	count := 0
	for rows.Next() {
		var genre domain.Genre
		if err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&genre.ParentID,
		); err != nil {
			return nil, 0, err
		}
		genres = append(genres, genre)
		count++
	}
	return genres, count, nil
}

func CreateBookGenre(ctx context.Context, conn *pgx.Conn, bookID, genreID int) error {
	sqlQuery := `
        INSERT INTO book_genres (book_id, genre_id)
        VALUES ($1, $2)
    `
	_, err := conn.Exec(ctx, sqlQuery, bookID, genreID)
	return err
}

func DeleteBookGenresByBookID(ctx context.Context, conn *pgx.Conn, bookID int) error {
	sqlQuery := `
        DELETE FROM book_genres
        WHERE book_id = $1
    `
	_, err := conn.Exec(ctx, sqlQuery, bookID)
	return err
}

func ExistsByNameExcludeIDGenre(ctx context.Context, conn *pgx.Conn, name string, excludeID int) (bool, error) {
	sqlQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM genres 
			WHERE name = $1 AND genres_id != $2
		)
	`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, name, excludeID).Scan(&exists)
	return exists, err
}

func CountSubGenres(ctx context.Context, conn *pgx.Conn, genreID int) (int, error) {
	sqlQuery := `
		SELECT COUNT(*)
		FROM genres
		WHERE parent_id = $1
	`
	var count int
	err := conn.QueryRow(ctx, sqlQuery, genreID).Scan(&count)
	return count, err
}

func CountBooksByGenreID(ctx context.Context, conn *pgx.Conn, genreID int) (int, error) {
	sqlQuery := `
		SELECT COUNT(*)
		FROM book_genres
		WHERE genre_id = $1
	`
	var count int
	err := conn.QueryRow(ctx, sqlQuery, genreID).Scan(&count)
	return count, err
}

func ExistsByNameGenre(ctx context.Context, conn *pgx.Conn, name string) (bool, error) {
	sqlQuery := `SELECT EXISTS (SELECT 1 FROM genres WHERE name = $1)`
	var exists bool
	err := conn.QueryRow(ctx, sqlQuery, name).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
