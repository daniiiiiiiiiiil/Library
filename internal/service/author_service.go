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
)

type AuthorService struct {
	authorRepo repository.AuthorRepository
	bookRepo   repository.BookRepository
	auditRepo  repository.AuditLogRepository
}

func NewAuthorService(
	authorRepo repository.AuthorRepository,
	bookRepo repository.BookRepository,
	auditRepo repository.AuditLogRepository,
) *AuthorService {
	return &AuthorService{
		authorRepo: authorRepo,
		bookRepo:   bookRepo,
		auditRepo:  auditRepo,
	}
}

func (s *AuthorService) CreateAuthor(ctx context.Context, conn *pgx.Conn, author *domain.Author) (*domain.Author, error) {
	if err := author.ValidateAuthor(); err != nil {
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации" + err.Error(),
		}
	}

	exists, err := s.authorRepo.ExistsByName(ctx, conn, author.First_name, author.Last_name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.BusinessError{
			Code:    "ErrAuthorAlreadyExists",
			Message: fmt.Sprintf("Автор с именем %s %s уже существует", author.First_name, author.Last_name),
		}
	}

	createAuthor, err := s.authorRepo.CreateAuthor(ctx, conn, author)
	if err != nil {
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
		return nil, err
	}

	return createAuthor, nil
}

func (s *AuthorService) GetAuthor(ctx context.Context, conn *pgx.Conn, id int) (*domain.Author, error) {
	if id <= 0 {
		return nil, errors.BusinessError{
			Code:    "ErrInvalidId",
			Message: "ID должен быть положительным числом",
		}
	}

	author, err := s.authorRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "Author",
			ID:     id,
		}
	}
	return &author, nil
}

func (s *AuthorService) UpdateAuthor(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}) (*domain.Author, error) {
	existingAuthor, err := s.authorRepo.GetByID(ctx, conn, id)
	if err != nil {
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
			return nil, errors.BusinessError{
				Code:    "ErrInvalidBirthdayFormat",
				Message: "Неверный формат даты рождения. Ожидается YYYY-MM-DD",
			}
		}
		existingAuthor.Birthday = parsedBirthday
	}

	if err := existingAuthor.ValidateAuthor(); err != nil {
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	exists, err := s.authorRepo.ExistsByNameExcludeID(ctx, conn, existingAuthor.First_name, existingAuthor.Last_name, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.BusinessError{
			Code:    "ErrAuthorAlreadyExists",
			Message: fmt.Sprintf("Автор с именем %s %s уже существует", existingAuthor.First_name, existingAuthor.Last_name),
		}
	}

	if err := s.authorRepo.Update(ctx, conn, existingAuthor); err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrUpdateAuthor",
			Message: "Не удалось обновить автора: " + err.Error(),
		}
	}

	return &existingAuthor, nil
}

func (s *AuthorService) DeleteAuthor(ctx context.Context, conn *pgx.Conn, id int) error {
	_, err := s.authorRepo.GetByID(ctx, conn, id)
	if err != nil {
		return errors.NotFoundError{
			Entity: "Author",
			ID:     id,
		}
	}

	books, err := s.authorRepo.GetByBookID(ctx, conn, id)
	if err != nil {
		return err
	}
	if len(books) > 0 {
		return errors.BusinessError{
			Code:    "ErrAuthorHasBooks",
			Message: fmt.Sprintf("Нельзя удалить автора, у него есть %d книг", len(books)),
		}
	}

	if err := s.authorRepo.Delete(ctx, conn, id); err != nil {
		return errors.BusinessError{
			Code:    "ErrDeleteAuthor",
			Message: "Не удалось удалить автора: " + err.Error(),
		}
	}

	return nil
}

func (s *AuthorService) ListAuthors(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Author, int, error) {
	limitOffset(limit, offset)

	authors, err := s.authorRepo.List(ctx, conn, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrListAuthorsError",
			Message: "Не удалось получить список авторов: " + err.Error(),
		}
	}

	total, err := s.authorRepo.CountAuthor(ctx, conn)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrCountAuthor",
			Message: "Не удалось подсчитать авторов: " + err.Error(),
		}
	}

	return authors, total, nil
}

func (s *AuthorService) GetAuthorsByBook(ctx context.Context, conn *pgx.Conn, bookID int) ([]domain.Author, error) {
	exists, err := s.bookRepo.Exists(ctx, conn, bookID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NotFoundError{
			Entity: "Book",
			ID:     bookID,
		}
	}

	authors, err := s.authorRepo.GetByBookID(ctx, conn, bookID)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "ErrGetByBookID",
			Message: "Не удалось получить авторов книги: " + err.Error(),
		}
	}

	return authors, nil
}

func (s *AuthorService) SearchAuthors(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Author, int, error) {
	limitOffset(limit, offset)
	if search == "" {
		return nil, 0, errors.BusinessError{
			Code:    "ErrEmptySearch",
			Message: "Поисковый запрос не может быть пустым",
		}
	}

	authors, count, err := s.authorRepo.Search(ctx, conn, column, search, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "ErrSearchAuthors",
			Message: "Ошибка при поиске авторов: " + err.Error(),
		}
	}

	return authors, count, nil
}
