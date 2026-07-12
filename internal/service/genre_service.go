package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"

	"github.com/jackc/pgx/v5"
)

type GenreService struct {
	genreRepo repository.GenreRepository
	bookRepo  repository.BookRepository
}

func NewGenreService(
	genreRepo repository.GenreRepository,
	bookRepo repository.BookRepository,
) *GenreService {
	return &GenreService{
		genreRepo: genreRepo,
		bookRepo:  bookRepo,
	}
}

func (g *GenreService) CreateGenre(ctx context.Context, conn *pgx.Conn, genre domain.Genre) (*domain.Genre, error) {
	if err := genre.ValidateGenre(); err != nil {
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации данных" + err.Error(),
		}
	}
	exists, err := g.genreRepo.ExistsByNameGenre(ctx, conn, genre.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.BusinessError{
			Code:    "ErrGenreAlreadyExists",
			Message: "Жанр с таким названием уже существует:" + err.Error(),
		}
	}

	if genre.ParentID != nil && *genre.ParentID > 0 {
		exists, err := g.genreRepo.Exists(ctx, conn, *genre.ParentID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.NotFoundError{
				Entity: "ParentGenre",
				ID:     *genre.ParentID,
			}
		}
	}

	genreCreate, err := g.genreRepo.CreateGenre(ctx, conn, genre)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrCreateGenre",
			Message: "Не удалось создать жанр:" + err.Error(),
		}
	}
	return genreCreate, nil
}

func (g *GenreService) GetGenre(ctx context.Context, conn *pgx.Conn, id int) (*domain.Genre, error) {
	if id < 0 {
		return nil, errors.ValidationError{
			Field:   "ID error validate",
			Message: "ID меньше 0",
		}
	}
	getGenre, err := g.genreRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Genre",
			ID:     id,
		}
	}
	return getGenre, nil
}

func (g *GenreService) UpdateGenre(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}) (*domain.Genre, error) {
	existingGenre, err := g.genreRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Genre",
			ID:     id,
		}
	}

	if name, ok := updates["name"].(string); ok {
		existingGenre.Name = name
	}
	if parentID, ok := updates["parent_id"].(int); ok {
		existingGenre.ParentID = &parentID
	}
	if parentID, ok := updates["parent_id"].(int); ok && parentID == 0 {
		existingGenre.ParentID = nil
	}

	if err := existingGenre.ValidateGenre(); err != nil {
		return nil, errors.BusinessError{
			Code:    "validation_error",
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	exists, err := g.genreRepo.ExistsByNameExcludeIDGenre(ctx, conn, existingGenre.Name, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.BusinessError{
			Code:    "genre_already_exists",
			Message: fmt.Sprintf("Жанр с названием '%s' уже существует", existingGenre.Name),
		}
	}

	if existingGenre.ParentID != nil && *existingGenre.ParentID > 0 {
		if *existingGenre.ParentID == id {
			return nil, errors.BusinessError{
				Code:    "invalid_parent_genre",
				Message: "Жанр не может быть родителем самого себя",
			}
		}

		exists, err := g.genreRepo.Exists(ctx, conn, *existingGenre.ParentID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.NotFoundError{
				Entity: "ParentGenre",
				ID:     *existingGenre.ParentID,
			}
		}
	}

	if err := g.genreRepo.Update(ctx, conn, *existingGenre); err != nil {
		return nil, errors.BusinessError{
			Code:    "update_genre_error",
			Message: "Не удалось обновить жанр: " + err.Error(),
		}
	}

	return existingGenre, nil
}

func (g *GenreService) DeleteGenre(ctx context.Context, conn *pgx.Conn, id int) (*domain.Genre, error) {
	exists, err := g.genreRepo.Exists(ctx, conn, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.BusinessError{
			Code:    "ErrGenreAlreadyExists",
			Message: "Не удалось проверить уникальность жанра:" + err.Error(),
		}
	}
	subGenresCount, err := g.genreRepo.CountSubGenres(ctx, conn, id)
	if err != nil {
		return nil, err
	}
	if subGenresCount > 0 {
		return nil, errors.BusinessError{
			Code:    "ErrGenreHasSubgenres",
			Message: fmt.Sprintf("Нельзя удалить жанр, у него есть %d поджанров", subGenresCount),
		}
	}

	booksCount, err := g.genreRepo.CountBooksByGenreID(ctx, conn, id)
	if err != nil {
		return nil, err
	}
	if booksCount > 0 {
		return nil, errors.BusinessError{
			Code:    "genre_has_books",
			Message: fmt.Sprintf("Нельзя удалить жанр, у него есть %d книг", booksCount),
		}
	}

	genreDelete, err := g.genreRepo.Delete(ctx, conn, id)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrDeleteGenre",
			Message: "Не удалось удалить жанр" + err.Error(),
		}
	}
	return genreDelete, nil
}

func (g *GenreService) ListGenres(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Genre, error) {
	limitOffset(limit, offset)
	genre, err := g.genreRepo.List(ctx, conn, limit, offset)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrGetListGenres",
			Message: "Не удалось получить список жанров" + err.Error(),
		}
	}
	return genre, nil
}

func (g *GenreService) GetSubGenres(ctx context.Context, conn *pgx.Conn, parentID int) ([]domain.Genre, error) {
	if parentID > 0 {
		exists, err := g.genreRepo.Exists(ctx, conn, parentID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.NotFoundError{
				Entity: "ParentGenre",
				ID:     parentID,
			}
		}
	}

	subGenre, err := g.genreRepo.GetSubGenres(ctx, conn, parentID)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrGetSubGenres",
			Message: "Не удалось получить поджанры" + err.Error(),
		}
	}
	return subGenre, nil
}

func (g *GenreService) GetRootGenres(ctx context.Context, conn *pgx.Conn) ([]domain.Genre, error) {
	genreRoot, err := g.genreRepo.GetRootGenres(ctx, conn)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrGetRootGenres",
			Message: "Не удалось получить корневые жанры" + err.Error(),
		}
	}
	return genreRoot, nil
}

func (g *GenreService) SearchGenres(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Genre, int, error) {
	limitOffset(limit, offset)
	if column == "" && search == "" && len(column) < 0 && len(search) < 0 {
		return nil, 0, errors.ValidationError{
			Field:   "ValidateErrorColumnOrSearch",
			Message: "Значения колонки и поиска пустые",
		}
	}
	resultSearch, count, err := g.genreRepo.Search(ctx, conn, column, search, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrGenrSearch",
			Message: "Не удалось сделать поиск по жанрам" + err.Error(),
		}
	}
	return resultSearch, count, nil
}
