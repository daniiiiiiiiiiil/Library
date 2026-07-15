package service

import (
	"context"
	"fmt"
	"library/internal/domain"
	"library/internal/infrastructure/audit"
	"library/internal/repository"
	"library/pkg/errors"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type BookService struct {
	BookRepo      repository.BookRepository
	CopyRepo      repository.BookCopyRepository
	AuthorRepo    repository.AuthorRepository
	GenreRepo     repository.GenreRepository
	PublisherRepo repository.PublisherRepository
	AuditRepo     repository.AuditLogRepository
	Logger        *zap.Logger
}

func NewBookService(
	bookRepo repository.BookRepository,
	copyRepo repository.BookCopyRepository,
	authorRepo repository.AuthorRepository,
	genreRepo repository.GenreRepository,
	publisherRepo repository.PublisherRepository,
	auditRepo repository.AuditLogRepository,
	logger *zap.Logger,
) *BookService {
	return &BookService{
		BookRepo:      bookRepo,
		CopyRepo:      copyRepo,
		AuthorRepo:    authorRepo,
		GenreRepo:     genreRepo,
		PublisherRepo: publisherRepo,
		AuditRepo:     auditRepo,
		Logger:        logger,
	}
}

func limitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func (b *BookService) CreateBook(ctx context.Context, conn *pgx.Conn, book *domain.Book, authorIDs []int, genreIDs []int) (*domain.Book, error) {
	b.Logger.Info("create book started", zap.String("isbn", book.ISBN), zap.Int("publisher_id", book.PublisherID), zap.Ints("author_ids", authorIDs), zap.Ints("genre_ids", genreIDs))

	if err := book.Validate(); err != nil {
		b.Logger.Warn("book validation failed", zap.String("isbn", book.ISBN), zap.Error(err))
		return nil, errors.ValidationError{
			Field:   err.Error(),
			Message: "Ошибка валидации: " + err.Error(),
		}
	}

	exists, err := b.BookRepo.ExistsByISBN(ctx, conn, book.ISBN)
	if err != nil {
		b.Logger.Error("failed to check book existence by isbn", zap.String("isbn", book.ISBN), zap.Error(err))
		return nil, err
	}
	if exists {
		b.Logger.Warn("book already exists", zap.String("isbn", book.ISBN))
		return nil, errors.BusinessError{
			Code:    "book_already_exists",
			Message: "книга с таким ISBN уже существует",
		}
	}

	if book.PublisherID > 0 {
		exists, err := b.PublisherRepo.Exists(ctx, conn, book.PublisherID)
		if err != nil {
			b.Logger.Error("failed to check publisher existence", zap.Int("publisher_id", book.PublisherID), zap.Error(err))
			return nil, err
		}
		if !exists {
			b.Logger.Warn("publisher not found", zap.Int("publisher_id", book.PublisherID))
			return nil, errors.BusinessError{
				Code:    "publisher_not_found",
				Message: "издатель не найден",
			}
		}
	}

	for _, authorID := range authorIDs {
		exists, err := b.AuthorRepo.Exists(ctx, conn, authorID)
		if err != nil {
			b.Logger.Error("failed to check author existence", zap.Int("author_id", authorID), zap.Error(err))
			return nil, err
		}
		if !exists {
			b.Logger.Warn("author not found", zap.Int("author_id", authorID))
			return nil, errors.BusinessError{
				Code:    "author_not_found",
				Message: fmt.Sprintf("автор с ID %d не найден", authorID),
			}
		}
	}

	for _, genreID := range genreIDs {
		exists, err := b.GenreRepo.Exists(ctx, conn, genreID)
		if err != nil {
			b.Logger.Error("failed to check genre existence", zap.Int("genre_id", genreID), zap.Error(err))
			return nil, err
		}
		if !exists {
			b.Logger.Warn("genre not found", zap.Int("genre_id", genreID))
			return nil, errors.BusinessError{
				Code:    "genre_not_found",
				Message: fmt.Sprintf("жанр с ID %d не найден", genreID),
			}
		}
	}

	newBook, err := b.BookRepo.CreateBook(ctx, conn, *book)
	if err != nil {
		b.Logger.Error("failed to create book", zap.String("isbn", book.ISBN), zap.Error(err))
		return nil, err
	}

	for _, authorID := range authorIDs {
		if err := b.AuthorRepo.CreateBookAuthor(ctx, conn, newBook.ID, authorID); err != nil {
			b.Logger.Error("failed to create book author relation", zap.Int("book_id", newBook.ID), zap.Int("author_id", authorID), zap.Error(err))
			return nil, err
		}
	}

	for _, genreID := range genreIDs {
		if err := b.GenreRepo.CreateBookGenre(ctx, conn, newBook.ID, genreID); err != nil {
			b.Logger.Error("failed to create book genre relation", zap.Int("book_id", newBook.ID), zap.Int("genre_id", genreID), zap.Error(err))
			return nil, err
		}
	}

	if err := b.AuditRepo.CreateAuditLog(ctx, conn, audit.AuditLog{
		UserID:     nil,
		Action:     "CREATE",
		EntityType: "book",
		EntityID:   newBook.ID,
	}); err != nil {
		b.Logger.Error("failed to create audit log for book create", zap.Int("book_id", newBook.ID), zap.Error(err))
		return nil, err
	}

	b.Logger.Info("book created successfully", zap.Int("book_id", newBook.ID), zap.String("isbn", newBook.ISBN))
	return newBook, nil
}

func (b *BookService) GetBook(ctx context.Context, conn *pgx.Conn, id int) (*domain.Book, error) {
	b.Logger.Debug("get book started", zap.Int("book_id", id))

	book, err := b.BookRepo.GetByID(ctx, conn, id)
	if err != nil {
		b.Logger.Warn("book not found", zap.Int("book_id", id), zap.Error(err))
		return nil, errors.NotFoundError{
			Entity: "book not found",
			ID:     id,
		}
	}

	if err := b.AuditRepo.CreateAuditLog(ctx, conn, audit.AuditLog{
		UserID:     nil,
		Action:     "GET",
		EntityType: "book",
		EntityID:   book.ID,
	}); err != nil {
		b.Logger.Error("failed to create audit log for get book", zap.Int("book_id", book.ID), zap.Error(err))
		return nil, err
	}

	b.Logger.Debug("get book finished", zap.Int("book_id", book.ID))
	return &book, nil
}

func (b *BookService) GetBookByISBN(ctx context.Context, conn *pgx.Conn, isbn string) (*domain.Book, error) {
	b.Logger.Debug("get book by isbn started", zap.String("isbn", isbn))

	book, err := b.BookRepo.GetByISBN(ctx, conn, isbn)
	if err != nil {
		b.Logger.Warn("book not found by isbn", zap.String("isbn", isbn), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "BookNotFound",
			Message: "Книги с таким isbn не существует: " + err.Error(),
		}
	}

	b.Logger.Debug("get book by isbn finished", zap.String("isbn", isbn), zap.Int("book_id", book.ID))
	return &book, nil
}

func (b *BookService) UpdateBook(ctx context.Context, conn *pgx.Conn, id int, updates map[string]interface{}, authorIDs []int, genreIDs []int) (*domain.Book, error) {
	b.Logger.Info("update book started", zap.Int("book_id", id), zap.Ints("author_ids", authorIDs), zap.Ints("genre_ids", genreIDs))

	existingBook, err := b.BookRepo.GetByID(ctx, conn, id)
	if err != nil {
		b.Logger.Warn("book not found for update", zap.Int("book_id", id), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "BookNotFound",
			Message: fmt.Sprintf("Книга с ID %d не найдена", id),
		}
	}

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

	if err := existingBook.Validate(); err != nil {
		b.Logger.Warn("book validation failed on update", zap.Int("book_id", id), zap.Error(err))
		return nil, errors.BusinessError{
			Code:    "validation_error",
			Message: err.Error(),
		}
	}

	if err := b.BookRepo.Update(ctx, conn, id, existingBook); err != nil {
		b.Logger.Error("failed to update book", zap.Int("book_id", id), zap.Error(err))
		return nil, err
	}

	if len(authorIDs) > 0 {
		if err := b.AuthorRepo.DeleteBookAuthorsByBookID(ctx, conn, id); err != nil {
			b.Logger.Error("failed to delete old book authors", zap.Int("book_id", id), zap.Error(err))
			return nil, err
		}
		for _, authorID := range authorIDs {
			if err := b.AuthorRepo.CreateBookAuthor(ctx, conn, id, authorID); err != nil {
				b.Logger.Error("failed to create updated book author relation", zap.Int("book_id", id), zap.Int("author_id", authorID), zap.Error(err))
				return nil, err
			}
		}
	}

	if len(genreIDs) > 0 {
		if err := b.GenreRepo.DeleteBookGenresByBookID(ctx, conn, id); err != nil {
			b.Logger.Error("failed to delete old book genres", zap.Int("book_id", id), zap.Error(err))
			return nil, err
		}
		for _, genreID := range genreIDs {
			if err := b.GenreRepo.CreateBookGenre(ctx, conn, id, genreID); err != nil {
				b.Logger.Error("failed to create updated book genre relation", zap.Int("book_id", id), zap.Int("genre_id", genreID), zap.Error(err))
				return nil, err
			}
		}
	}

	if err := b.AuditRepo.CreateAuditLog(ctx, conn, audit.AuditLog{
		UserID:     nil,
		Action:     "UPDATE",
		EntityType: "book",
		EntityID:   id,
	}); err != nil {
		b.Logger.Error("failed to create audit log for update book", zap.Int("book_id", id), zap.Error(err))
		return nil, err
	}

	b.Logger.Info("book updated successfully", zap.Int("book_id", id))
	return &existingBook, nil
}

func (b *BookService) DeleteBook(ctx context.Context, conn *pgx.Conn, bookID int, authorIDS []int, genreIDs []int) error {
	b.Logger.Info("delete book started", zap.Int("book_id", bookID))

	_, err := b.BookRepo.GetByID(ctx, conn, bookID)
	if err != nil {
		b.Logger.Warn("book not found for delete", zap.Int("book_id", bookID), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrBookNotFound",
			Message: fmt.Sprintf("Книга с ID %d не найдена", bookID),
		}
	}

	hasActiveCopies, err := b.CopyRepo.HasActiveCopies(ctx, conn, bookID)
	if err != nil {
		b.Logger.Error("failed to check active copies", zap.Int("book_id", bookID), zap.Error(err))
		return err
	}
	if hasActiveCopies {
		b.Logger.Warn("book has active copies", zap.Int("book_id", bookID))
		return errors.BusinessError{
			Code:    "ErrBookHasActiveCopies",
			Message: "Нельзя удалить книгу, у которой есть выданные или зарезервированные экземпляры",
		}
	}

	if err := b.BookRepo.Delete(ctx, conn, bookID); err != nil {
		b.Logger.Error("failed to delete book", zap.Int("book_id", bookID), zap.Error(err))
		return errors.BusinessError{
			Code:    "BookDeleteError",
			Message: "Не удалось удалить книгу: " + err.Error(),
		}
	}

	if len(authorIDS) > 0 {
		if err := b.AuthorRepo.DeleteBookAuthorsByBookID(ctx, conn, bookID); err != nil {
			b.Logger.Error("failed to delete book authors", zap.Int("book_id", bookID), zap.Error(err))
			return err
		}
	}

	if len(genreIDs) > 0 {
		if err := b.GenreRepo.DeleteBookGenresByBookID(ctx, conn, bookID); err != nil {
			b.Logger.Error("failed to delete book genres", zap.Int("book_id", bookID), zap.Error(err))
			return err
		}
	}

	if err := b.AuditRepo.CreateAuditLog(ctx, conn, audit.AuditLog{
		UserID:     nil,
		Action:     "DELETE",
		EntityType: "book",
		EntityID:   bookID,
	}); err != nil {
		b.Logger.Error("failed to create audit log for delete book", zap.Int("book_id", bookID), zap.Error(err))
		return err
	}

	b.Logger.Info("book deleted successfully", zap.Int("book_id", bookID))
	return nil
}

func (b *BookService) ListBooks(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, int, error) {
	limit, offset = limitOffset(limit, offset)
	b.Logger.Debug("list books started", zap.Int("limit", limit), zap.Int("offset", offset))

	books, err := b.BookRepo.List(ctx, conn, limit, offset)
	if err != nil {
		b.Logger.Error("failed to list books", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "list_books_error",
			Message: "Ошибка при получении списка книг: " + err.Error(),
		}
	}

	total, err := b.BookRepo.Count(ctx, conn)
	if err != nil {
		b.Logger.Error("failed to count books", zap.Error(err))
		return nil, 0, errors.BusinessError{
			Code:    "count_books_error",
			Message: "Ошибка при подсчёте книг: " + err.Error(),
		}
	}

	b.Logger.Debug("list books finished", zap.Int("returned", len(books)), zap.Int("total", total))
	return books, total, nil
}

func (b *BookService) SearchBooks(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Book, int, error) {
	limit, offset = limitOffset(limit, offset)
	b.Logger.Debug("search books started", zap.String("column", column), zap.String("search", search), zap.Int("limit", limit), zap.Int("offset", offset))

	books, count, err := b.BookRepo.Search(ctx, conn, column, search, limit, offset)
	if err != nil {
		b.Logger.Error("failed to search books", zap.String("column", column), zap.String("search", search), zap.Error(err))
		return []domain.Book{}, 0, errors.BusinessError{
			Code:    "search_books_error",
			Message: "Ошибка при поиске книг: " + err.Error(),
		}
	}

	b.Logger.Debug("search books finished", zap.Int("found", count))
	return books, count, nil
}

func (b *BookService) GetPopularBooks(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, error) {
	limit, offset = limitOffset(limit, offset)
	b.Logger.Debug("get popular books started", zap.Int("limit", limit), zap.Int("offset", offset))

	books, err := b.BookRepo.GetPopular(ctx, conn, limit, offset)
	if err != nil {
		b.Logger.Error("failed to get popular books", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return []domain.Book{}, err
	}

	b.Logger.Debug("get popular books finished", zap.Int("returned", len(books)))
	return books, nil
}

func (b *BookService) AddCopyToBook(ctx context.Context, conn *pgx.Conn, bookID int, condition string) error {
	b.Logger.Info("add copy to book started", zap.Int("book_id", bookID), zap.String("condition", condition))

	_, err := b.BookRepo.GetByID(ctx, conn, bookID)
	if err != nil {
		b.Logger.Warn("book not found for add copy", zap.Int("book_id", bookID), zap.Error(err))
		return errors.BusinessError{
			Code:    "book_not_found",
			Message: fmt.Sprintf("Книга с ID %d не найдена", bookID),
		}
	}

	nextNumber, err := b.CopyRepo.GetNextCopyNumber(ctx, conn, bookID)
	if err != nil {
		b.Logger.Error("failed to get next copy number", zap.Int("book_id", bookID), zap.Error(err))
		return err
	}

	newCopy := domain.BookCopy{
		BookID:     bookID,
		CopyNumber: nextNumber,
		Status:     "available",
		Condition:  condition,
		ReaderID:   nil,
		BorrowedAt: nil,
	}

	if err := newCopy.Validate(); err != nil {
		b.Logger.Warn("copy validation failed", zap.Int("book_id", bookID), zap.Error(err))
		return errors.BusinessError{
			Code:    "validation_error",
			Message: err.Error(),
		}
	}

	if err := b.CopyRepo.CreateCopy(ctx, conn, &newCopy); err != nil {
		b.Logger.Error("failed to create book copy", zap.Int("book_id", bookID), zap.Error(err))
		return errors.BusinessError{
			Code:    "ErrCreateBookCopy",
			Message: "Не удалось создать экземпляр книги: " + err.Error(),
		}
	}

	b.Logger.Info("copy created successfully", zap.Int("book_id", bookID), zap.Int("copy_number", newCopy.CopyNumber))
	return nil
}
