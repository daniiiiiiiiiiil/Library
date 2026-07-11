package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/infrastructure/audit"
	"library/internal/repository"
	"library/pkg/errors"

	"github.com/jackc/pgx/v5"
)

type BookService struct {
	BookRepo      repository.BookRepository
	CopyRepo      repository.BookCopyRepository
	AuthorRepo    repository.AuthorRepository
	GenreRepo     repository.GenreRepository
	PublisherRepo repository.PublisherRepository
	AuditRepo     repository.AuditLogRepository
}

func NewBookService(bookRepo repository.BookRepository, copyRepo repository.BookCopyRepository, authorRepo repository.AuthorRepository, genreRepo repository.GenreRepository, auditRepo repository.AuditLogRepository) *BookService {
	return &BookService{
		BookRepo:   bookRepo,
		CopyRepo:   copyRepo,
		AuthorRepo: authorRepo,
		GenreRepo:  genreRepo,
		AuditRepo:  auditRepo,
	}
}

func limitOffset(limit, offset int) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
}

func (b *BookService) CreateBook(ctx context.Context, conn *pgx.Conn, book *domain.Book, authorIDs []int, genreIDs []int) (*domain.Book, error) {
	// Валидация книги
	if err := book.Validate(); err != nil {
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации" + err.Error(),
		}
	}

	// Проверка уникальности ID
	exists, err := b.BookRepo.ExistsByISBN(ctx, conn, book.ISBN)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.BusinessError{
			Code:    "book_already_exists",
			Message: "книга с таким ID уже существует",
		}
	}

	// Проверка существования издателя
	if book.PublisherID > 0 {
		exists, err := b.PublisherRepo.Exists(ctx, conn, book.PublisherID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.BusinessError{
				Code:    "publisher_not_found",
				Message: "издатель не найден",
			}
		}
	}

	// Проверка существования авторов
	for _, authorID := range authorIDs {
		exists, err := b.AuthorRepo.Exists(ctx, conn, authorID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.BusinessError{
				Code:    "author_not_found",
				Message: "автор с ID " + string(rune(authorID)) + " не найден",
			}
		}
	}

	// Проверка существования жанров
	for _, genreID := range genreIDs {
		exists, err := b.GenreRepo.Exists(ctx, conn, genreID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.BusinessError{
				Code:    "genre_not_found",
				Message: "жанр с ID " + string(rune(genreID)) + " не найден",
			}
		}
	}

	//  Создание книги
	if err := b.BookRepo.Create(ctx, conn, *book); err != nil {
		return nil, err
	}

	//  Добавление связей с авторами
	for _, authorID := range authorIDs {
		if err := b.AuthorRepo.CreateBookAuthor(ctx, conn, book.ID, authorID); err != nil {
			return nil, err
		}
	}

	// Добавление связей с жанрами
	for _, genreID := range genreIDs {
		if err := b.GenreRepo.CreateBookGenre(ctx, conn, book.ID, genreID); err != nil {
			return nil, err
		}
	}

	// Логирование в аудит
	if err := b.AuditRepo.Create(ctx, conn, audit.AuditLog{
		UserID:     nil,
		Action:     "CREATE",
		EntityType: "book",
		EntityID:   book.ID,
	}); err != nil {
		return nil, err
	}

	// Возвращаем созданную книгу
	return book, nil
}

func (b *BookService) GetBook(ctx context.Context, conn *pgx.Conn, id int) (*domain.Book, error) {
	book, err := b.BookRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.NotFoundError{
			Entity: "book not found",
			ID:     id,
		}
	}
	if err := b.AuditRepo.Create(ctx, conn, audit.AuditLog{
		UserID:     nil,
		Action:     "GET",
		EntityType: "book",
		EntityID:   book.ID,
	}); err != nil {
		return nil, err
	}
	return &book, nil
}

func (b *BookService) GetBookByISBN(ctx context.Context, conn *pgx.Conn, isbn string) (*domain.Book, error) {
	book, err := b.BookRepo.GetByISBN(ctx, conn, isbn)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "BookNotFound",
			Message: "Книги с таким isbn не существует:" + err.Error(),
		}
	}
	return &book, nil
}

func (b *BookService) UpdateBook(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}, authorIDs []int, genreIDs []int) (*domain.Book, error) {
	//  Проверяем, что книга существует
	existingBook, err := b.BookRepo.GetByID(ctx, conn, id)
	if err != nil {
		return nil, errors.BusinessError{
			Code:    "BookNotFound",
			Message: fmt.Sprintf("Книга с ID %d не найдена", id),
		}
	}

	//  Применяем обновления к книге
	if title, ok := updates["title"].(string); ok {
		existingBook.Title = title
	}
	if isbn, ok := updates["isbn"].(string); ok {
		existingBook.ISBN = isbn
	}
	if year, ok := updates["year"].(int); ok {
		existingBook.Year = year
	}
	if publisherID, ok := updates["publisher_id"].(int); ok {
		existingBook.PublisherID = publisherID
	}
	if description, ok := updates["description"].(string); ok {
		existingBook.Description = description
	}
	if coverImageURL, ok := updates["cover_image_url"].(string); ok {
		existingBook.CoverImageURL = coverImageURL
	}

	// Валидация обновленной книги
	if err := existingBook.Validate(); err != nil {
		return nil, errors.BusinessError{
			Code:    "validation_error",
			Message: err.Error(),
		}
	}

	// Обновляем книгу в БД
	if err := b.BookRepo.Update(ctx, conn, id, existingBook); err != nil {
		return nil, err
	}

	// Обновляем связи с авторами (если переданы)
	if len(authorIDs) > 0 {
		// Удаляем старые связи
		if err := b.AuthorRepo.DeleteBookAuthorsByBookID(ctx, conn, id); err != nil {
			return nil, err
		}
		// Добавляем новые
		for _, authorID := range authorIDs {
			if err := b.AuthorRepo.CreateBookAuthor(ctx, conn, id, authorID); err != nil {
				return nil, err
			}
		}
	}

	// Обновляем связи с жанрами (если переданы)
	if len(genreIDs) > 0 {
		// Удаляем старые связи
		if err := b.GenreRepo.DeleteBookGenresByBookID(ctx, conn, id); err != nil {
			return nil, err
		}
		// Добавляем новые
		for _, genreID := range genreIDs {
			if err := b.GenreRepo.CreateBookGenre(ctx, conn, id, genreID); err != nil {
				return nil, err
			}
		}
	}

	// Логирование в аудит
	if err := b.AuditRepo.Create(ctx, conn, audit.AuditLog{
		UserID:     nil,
		Action:     "UPDATE",
		EntityType: "book",
		EntityID:   id,
	}); err != nil {
		return nil, err
	}

	return &existingBook, nil
}

func (b *BookService) DeleteBook(ctx context.Context, conn *pgx.Conn, bookID int, authorIDS []int, genreIDs []int) error {
	_, err := b.BookRepo.GetByID(ctx, conn, bookID)
	if err != nil {
		return errors.BusinessError{
			Code:    "ErrBookNotFound",
			Message: fmt.Sprintf("Книга с ID %d не найдена", bookID),
		}
	}
	hasActiveCopies, err := b.CopyRepo.HasActiveCopies(ctx, conn, bookID)
	if err != nil {
		return err
	}
	if hasActiveCopies {
		return errors.BusinessError{
			Code:    "ErrBookHasActiveCopies",
			Message: "Нельзя удалить книгу, у которой есть выданные или зарезервированные экземпляры",
		}
	}
	if err := b.BookRepo.Delete(ctx, conn, bookID); err != nil {
		return errors.BusinessError{
			Code:    "BookDeleteError",
			Message: "Не удалось удалить книгу:" + err.Error(),
		}
	}
	if len(authorIDS) > 0 {
		if err := b.AuthorRepo.DeleteBookAuthorsByBookID(ctx, conn, bookID); err != nil {
			return err
		}
	}

	if len(genreIDs) > 0 {
		if err := b.GenreRepo.DeleteBookGenresByBookID(ctx, conn, bookID); err != nil {
			return err
		}
	}

	if err := b.AuditRepo.Create(ctx, conn, audit.AuditLog{
		UserID:     nil,
		Action:     "DELETE",
		EntityType: "book",
		EntityID:   bookID,
	}); err != nil {
		return err
	}
	return nil
}

func (b *BookService) ListBooks(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, int, error) {
	limitOffset(limit, offset)
	books, err := b.BookRepo.List(ctx, conn, limit, offset)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "list_books_error",
			Message: "Ошибка при получении списка книг: " + err.Error(),
		}
	}

	total, err := b.BookRepo.Count(ctx, conn)
	if err != nil {
		return nil, 0, errors.BusinessError{
			Code:    "count_books_error",
			Message: "Ошибка при подсчёте книг: " + err.Error(),
		}
	}
	return books, total, nil
}

func (b *BookService) SearchBooks(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Book, int, error) {
	limitOffset(limit, offset)
	book, count, err := b.BookRepo.Search(ctx, conn, column, search, limit, offset)
	if err != nil {
		return []domain.Book{}, 0, errors.BusinessError{
			Code:    "search_books_error",
			Message: "Ошибка при поиске книг: " + err.Error(),
		}
	}
	return book, count, nil
}

func (b *BookService) GetPopularBooks(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, error) {
	limitOffset(limit, offset)
	book, err := b.BookRepo.GetPopular(ctx, conn, limit, offset)
	if err != nil {
		return []domain.Book{}, err
	}
	return book, nil
}

func (b *BookService) AddCopyToBook(ctx context.Context, conn *pgx.Conn, bookID int, condition string) error {
	// Проверяем, что книга существует
	_, err := b.BookRepo.GetByID(ctx, conn, bookID)
	if err != nil {
		return errors.BusinessError{
			Code:    "book_not_found",
			Message: fmt.Sprintf("Книга с ID %d не найдена", bookID),
		}
	}

	// Получаем следующий номер копии
	nextNumber, err := b.CopyRepo.GetNextCopyNumber(ctx, conn, bookID)
	if err != nil {
		return err
	}

	// Создаём новую копию
	newCopy := domain.BookCopy{
		BookID:     bookID,
		CopyNumber: nextNumber,
		Status:     "available",
		Condition:  condition,
		ReaderID:   nil,
		BorrowedAt: nil,
	}

	if err := newCopy.Validate(); err != nil {
		return errors.BusinessError{
			Code:    "validation_error",
			Message: err.Error(),
		}
	}

	if err := b.CopyRepo.Create(ctx, conn, newCopy); err != nil {
		return errors.BusinessError{
			Code:    "create_copy_error",
			Message: "Не удалось создать экземпляр книги: " + err.Error(),
		}
	}

	return nil
}
