package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/repository"
	"library/pkg/errors"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type GenreService struct {
	genreRepo repository.GenreRepository
	bookRepo  repository.BookRepository
	logger    *zap.Logger
}

func NewGenreService(
	genreRepo repository.GenreRepository,
	bookRepo repository.BookRepository,
	logger *zap.Logger,
) *GenreService {
	return &GenreService{
		genreRepo: genreRepo,
		bookRepo:  bookRepo,
		logger:    logger,
	}
}

func (g *GenreService) CreateGenre(ctx context.Context, conn *pgx.Conn, genre domain.Genre) (*domain.Genre, error) {
	g.logger.Info("create genre started", zap.String("name", genre.Name), zap.Intp("parent_id", genre.ParentID))

	if err := genre.ValidateGenre(); err != nil {
		g.logger.Warn("genre validation failed", zap.String("name", genre.Name), zap.Error(err))
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации данных" + err.Error(),
		}
	}

	exists, err := g.genreRepo.ExistsByName(ctx, conn, genre.Name)
	if err != nil {
		g.logger.Error("failed to check genre existence by name", zap.String("name", genre.Name), zap.Error(err))
		return nil, err
	}
	if exists {
		g.logger.Warn("genre already exists", zap.String("name", genre.Name))
		return nil, errors.BusinessError{
			Code:    "ErrGenreAlreadyExists",
			Message: fmt.Sprintf("Жанр с названием '%s' уже существует", genre.Name), // <-- ИСПРАВЛЕНО
		}
	}

	if genre.ParentID != nil && *genre.ParentID > 0 {
		exists, err := g.genreRepo.Exists(ctx, conn, *genre.ParentID)
		if err != nil {
			g.logger.Error("failed to check parent genre existence", zap.Int("parent_id", *genre.ParentID), zap.Error(err))
			return nil, err
		}
		if !exists {
			g.logger.Warn("parent genre not found", zap.Int("parent_id", *genre.ParentID))
			return nil, errors.NotFoundError{
				Entity: "ParentGenre",
				ID:     *genre.ParentID,
			}
		}
	}

	genreCreate, err := g.genreRepo.CreateGenre(ctx, conn, genre)
	if err != nil {
		g.logger.Error("failed to create genre", zap.String("name", genre.Name), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrCreateGenre",
			Message: "Не удалось создать жанр:" + err.Error(),
		}
	}

	g.logger.Info("genre created successfully", zap.Int("genre_id", genreCreate.ID), zap.String("name", genreCreate.Name))
	return genreCreate, nil
}

func (g *GenreService) GetGenre(ctx context.Context, conn *pgx.Conn, id int) (*domain.Genre, error) {
	g.logger.Debug("get genre started", zap.Int("genre_id", id))

	if id < 0 {
		g.logger.Warn("invalid genre id", zap.Int("genre_id", id))
		return nil, errors.ValidationError{
			Field:   "ID error validate",
			Message: "ID меньше 0",
		}
	}
	getGenre, err := g.genreRepo.GetByID(ctx, conn, id)
	if err != nil {
		g.logger.Warn("genre not found", zap.Int("genre_id", id), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Genre",
			ID:     id,
		}
	}

	g.logger.Debug("get genre finished", zap.Int("genre_id", getGenre.ID))
	return getGenre, nil
}

func (g *GenreService) UpdateGenre(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}) (*domain.Genre, error) {
	g.logger.Info("update genre started", zap.Int("genre_id", id))

	existingGenre, err := g.genreRepo.GetByID(ctx, conn, id)
	if err != nil {
		g.logger.Warn("genre not found for update", zap.Int("genre_id", id), zap.Error(err))
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
		g.logger.Warn("genre validation failed on update", zap.Int("genre_id", id), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "validation_error",
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	exists, err := g.genreRepo.ExistsByNameExcludeID(ctx, conn, existingGenre.Name, id)
	if err != nil {
		g.logger.Error("failed to check genre existence by name exclude id", zap.Int("genre_id", id), zap.String("name", existingGenre.Name), zap.Error(err))
		return nil, err
	}
	if exists {
		g.logger.Warn("genre already exists with same name", zap.Int("genre_id", id), zap.String("name", existingGenre.Name))
		return nil, errors.BusinessError{
			Code:    "genre_already_exists",
			Message: fmt.Sprintf("Жанр с названием '%s' уже существует", existingGenre.Name),
		}
	}

	if existingGenre.ParentID != nil && *existingGenre.ParentID > 0 {
		if *existingGenre.ParentID == id {
			g.logger.Warn("genre cannot be parent of itself", zap.Int("genre_id", id))
			return nil, errors.BusinessError{
				Code:    "invalid_parent_genre",
				Message: "Жанр не может быть родителем самого себя",
			}
		}

		exists, err := g.genreRepo.Exists(ctx, conn, *existingGenre.ParentID)
		if err != nil {
			g.logger.Error("failed to check parent genre existence", zap.Int("parent_id", *existingGenre.ParentID), zap.Error(err))
			return nil, err
		}
		if !exists {
			g.logger.Warn("parent genre not found", zap.Int("parent_id", *existingGenre.ParentID))
			return nil, errors.NotFoundError{
				Entity: "ParentGenre",
				ID:     *existingGenre.ParentID,
			}
		}
	}

	if err := g.genreRepo.Update(ctx, conn, *existingGenre); err != nil {
		g.logger.Error("failed to update genre", zap.Int("genre_id", id), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "update_genre_error",
			Message: "Не удалось обновить жанр: " + err.Error(),
		}
	}

	g.logger.Info("genre updated successfully", zap.Int("genre_id", id))
	return existingGenre, nil
}

func (g *GenreService) DeleteGenre(ctx context.Context, conn *pgx.Conn, id int) (*domain.Genre, error) {
	g.logger.Info("delete genre started", zap.Int("genre_id", id))

	exists, err := g.genreRepo.Exists(ctx, conn, id)
	if err != nil {
		g.logger.Error("failed to check genre existence", zap.Int("genre_id", id), zap.Error(err))
		return nil, err
	}
	if !exists {
		g.logger.Warn("genre not found for delete", zap.Int("genre_id", id))
		return nil, errors.BusinessError{
			Code:    "ErrGenreAlreadyExists",
			Message: "Не удалось проверить уникальность жанра:" + err.Error(),
		}
	}
	subGenresCount, err := g.genreRepo.CountSubGenres(ctx, conn, id)
	if err != nil {
		g.logger.Error("failed to count subgenres", zap.Int("genre_id", id), zap.Error(err))
		return nil, err
	}
	if subGenresCount > 0 {
		g.logger.Warn("genre has subgenres, cannot delete", zap.Int("genre_id", id), zap.Int("subgenres_count", subGenresCount))
		return nil, errors.BusinessError{
			Code:    "ErrGenreHasSubgenres",
			Message: fmt.Sprintf("Нельзя удалить жанр, у него есть %d поджанров", subGenresCount),
		}
	}

	booksCount, err := g.genreRepo.CountBooksByGenreID(ctx, conn, id)
	if err != nil {
		g.logger.Error("failed to count books by genre", zap.Int("genre_id", id), zap.Error(err))
		return nil, err
	}
	if booksCount > 0 {
		g.logger.Warn("genre has books, cannot delete", zap.Int("genre_id", id), zap.Int("books_count", booksCount))
		return nil, errors.BusinessError{
			Code:    "genre_has_books",
			Message: fmt.Sprintf("Нельзя удалить жанр, у него есть %d книг", booksCount),
		}
	}

	genreDelete, err := g.genreRepo.Delete(ctx, conn, id)
	if err != nil {
		g.logger.Error("failed to delete genre", zap.Int("genre_id", id), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrDeleteGenre",
			Message: "Не удалось удалить жанр" + err.Error(),
		}
	}

	g.logger.Info("genre deleted successfully", zap.Int("genre_id", id))
	return genreDelete, nil
}

func (g *GenreService) ListGenres(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Genre, error) {
	limit, offset = limitOffset(limit, offset)
	g.logger.Debug("list genres started", zap.Int("limit", limit), zap.Int("offset", offset))

	genre, err := g.genreRepo.List(ctx, conn, limit, offset)
	if err != nil {
		g.logger.Error("failed to list genres", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrGetListGenres",
			Message: "Не удалось получить список жанров" + err.Error(),
		}
	}

	g.logger.Debug("list genres finished", zap.Int("returned", len(genre)))
	return genre, nil
}

func (g *GenreService) GetSubGenres(ctx context.Context, conn *pgx.Conn, parentID int) ([]domain.Genre, error) {
	g.logger.Debug("get subgenres started", zap.Int("parent_id", parentID))

	if parentID > 0 {
		exists, err := g.genreRepo.Exists(ctx, conn, parentID)
		if err != nil {
			g.logger.Error("failed to check parent genre existence", zap.Int("parent_id", parentID), zap.Error(err))
			return nil, err
		}
		if !exists {
			g.logger.Warn("parent genre not found", zap.Int("parent_id", parentID))
			return nil, errors.NotFoundError{
				Entity: "ParentGenre",
				ID:     parentID,
			}
		}
	}

	subGenre, err := g.genreRepo.GetSubGenres(ctx, conn, parentID)
	if err != nil {
		g.logger.Error("failed to get subgenres", zap.Int("parent_id", parentID), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrGetSubGenres",
			Message: "Не удалось получить поджанры" + err.Error(),
		}
	}

	g.logger.Debug("get subgenres finished", zap.Int("parent_id", parentID), zap.Int("subgenres_count", len(subGenre)))
	return subGenre, nil
}

func (g *GenreService) GetRootGenres(ctx context.Context, conn *pgx.Conn) ([]domain.Genre, error) {
	g.logger.Debug("get root genres started")

	genreRoot, err := g.genreRepo.GetRootGenres(ctx, conn)
	if err != nil {
		g.logger.Error("failed to get root genres", zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrGetRootGenres",
			Message: "Не удалось получить корневые жанры" + err.Error(),
		}
	}

	g.logger.Debug("get root genres finished", zap.Int("root_genres_count", len(genreRoot)))
	return genreRoot, nil
}

func (g *GenreService) SearchGenres(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Genre, int, error) {
	limit, offset = limitOffset(limit, offset)
	g.logger.Debug("search genres started", zap.String("column", column), zap.String("search", search), zap.Int("limit", limit), zap.Int("offset", offset))

	if column == "" && search == "" && len(column) < 0 && len(search) < 0 {
		g.logger.Warn("empty search parameters")
		return nil, 0, errors.ValidationError{
			Field:   "ValidateErrorColumnOrSearch",
			Message: "Значения колонки и поиска пустые",
		}
	}
	resultSearch, count, err := g.genreRepo.Search(ctx, conn, column, search, limit, offset)
	if err != nil {
		g.logger.Error("failed to search genres", zap.String("column", column), zap.String("search", search), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrGenrSearch",
			Message: "Не удалось сделать поиск по жанрам" + err.Error(),
		}
	}

	g.logger.Debug("search genres finished", zap.Int("found", count))
	return resultSearch, count, nil
}
