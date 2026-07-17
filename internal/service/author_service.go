package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/infrastructure/audit"
	"library/internal/repository"
	"library/pkg/errors"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type AuthorService struct {
	authorRepo repository.AuthorRepository
	bookRepo   repository.BookRepository
	auditRepo  repository.AuditLogRepository
	logger     *zap.Logger
}

func NewAuthorService(
	authorRepo repository.AuthorRepository,
	bookRepo repository.BookRepository,
	auditRepo repository.AuditLogRepository,
	logger *zap.Logger,
) *AuthorService {
	return &AuthorService{
		authorRepo: authorRepo,
		bookRepo:   bookRepo,
		auditRepo:  auditRepo,
		logger:     logger,
	}
}
func (s *AuthorService) CreateAuthor(ctx context.Context, conn *pgx.Conn, author *domain.Author) (*domain.Author, error) {
	s.logger.Info("create author started", zap.String("first_name", author.First_name), zap.String("last_name", author.Last_name))

	if err := author.ValidateAuthor(); err != nil {
		s.logger.Warn("author validation failed", zap.String("first_name", author.First_name), zap.String("last_name", author.Last_name), zap.Error(err))
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации" + err.Error(),
		}
	}

	exists, err := s.authorRepo.ExistsByName(ctx, conn, author.First_name, author.Last_name)
	if err != nil {
		s.logger.Error("failed to check author existence by name", zap.String("first_name", author.First_name), zap.String("last_name", author.Last_name), zap.Error(err))
		return nil, err
	}
	if exists {
		s.logger.Warn("author already exists", zap.String("first_name", author.First_name), zap.String("last_name", author.Last_name))
		return nil, errors.BusinessError{
			Code:    "ErrAuthorAlreadyExists",
			Message: fmt.Sprintf("Автор с именем %s %s уже существует", author.First_name, author.Last_name),
		}
	}

	createAuthor, err := s.authorRepo.CreateAuthor(ctx, conn, author)
	if err != nil {
		s.logger.Error("failed to create author", zap.String("first_name", author.First_name), zap.String("last_name", author.Last_name), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrCreateAuthor",
			Message: "Не удалось создать автора: " + err.Error(),
		}
	}

	if err := s.auditRepo.CreateAuditLog(ctx, conn, audit.AuditLog{
		UserID:     nil,
		Action:     "CREATE",
		EntityType: "author",
		EntityID:   author.ID,
	}); err != nil {
		s.logger.Error("failed to create audit log for author create", zap.Int("author_id", author.ID), zap.Error(err))
		return nil, err
	}

	s.logger.Info("author created successfully", zap.Int("author_id", createAuthor.ID), zap.String("first_name", createAuthor.First_name), zap.String("last_name", createAuthor.Last_name))
	return createAuthor, nil
}

func (s *AuthorService) GetAuthor(ctx context.Context, conn *pgx.Conn, id int) (*domain.Author, error) {
	s.logger.Debug("get author started", zap.Int("author_id", id))

	if id <= 0 {
		s.logger.Warn("invalid author id", zap.Int("author_id", id))
		return nil, errors.BusinessError{
			Code:    "ErrInvalidId",
			Message: "ID должен быть положительным числом",
		}
	}

	author, err := s.authorRepo.GetByID(ctx, conn, id)
	if err != nil {
		s.logger.Warn("author not found", zap.Int("author_id", id), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Author",
			ID:     id,
		}
	}

	s.logger.Debug("get author finished", zap.Int("author_id", author.ID))
	return &author, nil
}

func (s *AuthorService) UpdateAuthor(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}) (*domain.Author, error) {
	s.logger.Info("update author started", zap.Int("author_id", id))

	existingAuthor, err := s.authorRepo.GetByID(ctx, conn, id)
	if err != nil {
		s.logger.Warn("author not found for update", zap.Int("author_id", id), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "Author",
			ID:     id,
		}
	}

	if firstName, ok := updates["first_name"].(string); ok {
		existingAuthor.First_name = firstName
	}
	if lastName, ok := updates["last_name"].(string); ok {
		existingAuthor.Last_name = lastName
	}
	if biography, ok := updates["biography"].(string); ok {
		existingAuthor.Biography = biography
	}
	if birthday, ok := updates["birthday"].(string); ok {
		parsedBirthday, err := time.Parse("2006-01-02", birthday)
		if err != nil {
			s.logger.Warn("invalid birthday format", zap.Int("author_id", id), zap.String("birthday", birthday), zap.Error(err))
			return nil, errors.BusinessError{
				Code:    "ErrInvalidBirthdayFormat",
				Message: "Неверный формат даты рождения. Ожидается YYYY-MM-DD",
			}
		}
		existingAuthor.Birthday = parsedBirthday
	}

	if err := existingAuthor.ValidateAuthor(); err != nil {
		s.logger.Warn("author validation failed on update", zap.Int("author_id", id), zap.Error(err))
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	exists, err := s.authorRepo.ExistsByNameExcludeID(ctx, conn, existingAuthor.First_name, existingAuthor.Last_name, id)
	if err != nil {
		s.logger.Error("failed to check author existence by name exclude id", zap.Int("author_id", id), zap.String("first_name", existingAuthor.First_name), zap.String("last_name", existingAuthor.Last_name), zap.Error(err))
		return nil, err
	}
	if exists {
		s.logger.Warn("author already exists with same name", zap.Int("author_id", id), zap.String("first_name", existingAuthor.First_name), zap.String("last_name", existingAuthor.Last_name))
		return nil, errors.BusinessError{
			Code:    "ErrAuthorAlreadyExists",
			Message: fmt.Sprintf("Автор с именем %s %s уже существует", existingAuthor.First_name, existingAuthor.Last_name),
		}
	}

	if err := s.authorRepo.Update(ctx, conn, existingAuthor); err != nil {
		s.logger.Error("failed to update author", zap.Int("author_id", id), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "ErrUpdateAuthor",
			Message: "Не удалось обновить автора: " + err.Error(),
		}
	}

	s.logger.Info("author updated successfully", zap.Int("author_id", id))
	return &existingAuthor, nil
}

func (s *AuthorService) DeleteAuthor(ctx context.Context, conn *pgx.Conn, id int) error {
	s.logger.Info("delete author started", zap.Int("author_id", id))

	_, err := s.authorRepo.GetByID(ctx, conn, id)
	if err != nil {
		s.logger.Warn("author not found for delete", zap.Int("author_id", id), zap.Error(err))
		return errors.NotFoundError{
			Entity: "Author",
			ID:     id,
		}
	}

	books, err := s.authorRepo.GetByBookID(ctx, conn, id, 0, 0)
	if err != nil {
		s.logger.Error("failed to get author books", zap.Int("author_id", id), zap.Error(err))
		return err
	}
	if len(books) > 0 {
		s.logger.Warn("author has books, cannot delete", zap.Int("author_id", id), zap.Int("book_count", len(books)))
		return errors.BusinessError{
			Code:    "ErrAuthorHasBooks",
			Message: fmt.Sprintf("Нельзя удалить автора, у него есть %d книг", len(books)),
		}
	}

	if err := s.authorRepo.Delete(ctx, conn, id); err != nil {
		s.logger.Error("failed to delete author", zap.Int("author_id", id), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrDeleteAuthor",
			Message: "Не удалось удалить автора: " + err.Error(),
		}
	}

	s.logger.Info("author deleted successfully", zap.Int("author_id", id))
	return nil
}

func (s *AuthorService) ListAuthors(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Author, int, error) {
	limit, offset = limitOffset(limit, offset)
	s.logger.Debug("list authors started", zap.Int("limit", limit), zap.Int("offset", offset))

	authors, err := s.authorRepo.List(ctx, conn, limit, offset)
	if err != nil {
		s.logger.Error("failed to list authors", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrListAuthorsError",
			Message: "Не удалось получить список авторов: " + err.Error(),
		}
	}

	total, err := s.authorRepo.CountAuthor(ctx, conn)
	if err != nil {
		s.logger.Error("failed to count authors", zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrCountAuthor",
			Message: "Не удалось подсчитать авторов: " + err.Error(),
		}
	}

	s.logger.Debug("list authors finished", zap.Int("returned", len(authors)), zap.Int("total", total))
	return authors, total, nil
}

func (s *AuthorService) GetAuthorsByBook(ctx context.Context, conn *pgx.Conn, bookID int, limit, offset int) ([]domain.Author, int, error) {
	s.logger.Debug("get authors by book started", zap.Int("book_id", bookID))

	exists, err := s.bookRepo.Exists(ctx, conn, bookID)
	if err != nil {
		s.logger.Error("failed to check book existence", zap.Int("book_id", bookID), zap.Error(err))
		return nil, 0, err
	}
	if !exists {
		s.logger.Warn("book not found for get authors", zap.Int("book_id", bookID))
		return nil, 0, errors.NotFoundError{
			Entity: "Book",
			ID:     bookID,
		}
	}

	authors, err := s.authorRepo.GetByBookID(ctx, conn, bookID, 0, 0)
	if err != nil {
		s.logger.Error("failed to get authors by book id", zap.Int("book_id", bookID), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrGetByBookID",
			Message: "Не удалось получить авторов книги: " + err.Error(),
		}
	}

	s.logger.Debug("get authors by book finished", zap.Int("book_id", bookID), zap.Int("author_count", len(authors)))
	return authors, len(authors), nil
}

func (s *AuthorService) SearchAuthors(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Author, int, error) {
	limit, offset = limitOffset(limit, offset)
	s.logger.Debug("search authors started", zap.String("column", column), zap.String("search", search), zap.Int("limit", limit), zap.Int("offset", offset))

	if search == "" {
		s.logger.Warn("empty search query")
		return nil, 0, errors.BusinessError{
			Code:    "ErrEmptySearch",
			Message: "Поисковый запрос не может быть пустым",
		}
	}

	authors, count, err := s.authorRepo.Search(ctx, conn, column, search, limit, offset)
	if err != nil {
		s.logger.Error("failed to search authors", zap.String("column", column), zap.String("search", search), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "ErrSearchAuthors",
			Message: "Ошибка при поиске авторов: " + err.Error(),
		}
	}

	s.logger.Debug("search authors finished", zap.Int("found", count))
	return authors, count, nil
}
