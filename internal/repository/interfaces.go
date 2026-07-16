package repository

import (
	"context"
	"library/internal/domain"
	"library/internal/infrastructure/audit"
	"time"

	"github.com/jackc/pgx/v5"
)

// BookRepository определяет методы для работы с книгами
type BookRepository interface {
	CreateBook(ctx context.Context, conn *pgx.Conn, book domain.Book) (*domain.Book, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Book, error)
	GetByISBN(ctx context.Context, conn *pgx.Conn, isbn string) (domain.Book, error)
	Update(ctx context.Context, conn *pgx.Conn, id int, book domain.Book) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, error)
	Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Book, int, error)
	GetPopular(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, error)
	Exists(ctx context.Context, conn *pgx.Conn, bookID int) (bool, error)
	ExistsByISBN(ctx context.Context, conn *pgx.Conn, isbn string) (bool, error)
	Count(ctx context.Context, conn *pgx.Conn) (int, error)
	GetByPublisherID(ctx context.Context, conn *pgx.Conn, publisherID, limit, offset int) ([]domain.Book, error)
	CountByPublisherID(ctx context.Context, conn *pgx.Conn, publisherID int) (int, error)
	UpdateRating(ctx context.Context, conn *pgx.Conn, bookID int, avgRating float64) error
	UpdateRatingAndCount(ctx context.Context, conn *pgx.Conn, bookID int, reviewsCount int) error
}

// AuthorRepository определяет методы для работы с авторами
type AuthorRepository interface {
	CreateAuthor(ctx context.Context, conn *pgx.Conn, author *domain.Author) (*domain.Author, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Author, error)
	Update(ctx context.Context, conn *pgx.Conn, author domain.Author) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Author, error)
	Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Author, int, error)
	Exists(ctx context.Context, conn *pgx.Conn, authorID int) (bool, error)
	ExistsByName(ctx context.Context, conn *pgx.Conn, firstName, lastName string) (bool, error)
	ExistsByNameExcludeID(ctx context.Context, conn *pgx.Conn, firstName, lastName string, excludeID int) (bool, error)
	GetByBookID(ctx context.Context, conn *pgx.Conn, bookID int) ([]domain.Author, error)
	CreateBookAuthor(ctx context.Context, conn *pgx.Conn, bookID, authorID int) error
	DeleteBookAuthorsByBookID(ctx context.Context, conn *pgx.Conn, bookID int) error
	CountAuthor(ctx context.Context, conn *pgx.Conn) (int, error)
	GetBooksByAuthorID(ctx context.Context, conn *pgx.Conn, authorID int) ([]domain.Book, error)
}

// GenreRepository определяет методы для работы с жанрами
type GenreRepository interface {
	CreateGenre(ctx context.Context, conn *pgx.Conn, genre domain.Genre) (*domain.Genre, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.Genre, error)
	Update(ctx context.Context, conn *pgx.Conn, genre domain.Genre) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) (*domain.Genre, error)
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Genre, error)
	Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Genre, int, error)
	Exists(ctx context.Context, conn *pgx.Conn, id int) (bool, error)
	ExistsByName(ctx context.Context, conn *pgx.Conn, name string) (bool, error)
	ExistsByNameExcludeID(ctx context.Context, conn *pgx.Conn, name string, excludeID int) (bool, error)
	GetSubGenres(ctx context.Context, conn *pgx.Conn, parentID int) ([]domain.Genre, error)
	GetRootGenres(ctx context.Context, conn *pgx.Conn) ([]domain.Genre, error)
	CreateBookGenre(ctx context.Context, conn *pgx.Conn, bookID, genreID int) error
	DeleteBookGenresByBookID(ctx context.Context, conn *pgx.Conn, bookID int) error
	CountSubGenres(ctx context.Context, conn *pgx.Conn, genreID int) (int, error)
	CountBooksByGenreID(ctx context.Context, conn *pgx.Conn, genreID int) (int, error)
}

// PublisherRepository определяет методы для работы с издательствами
type PublisherRepository interface {
	Create(ctx context.Context, conn *pgx.Conn, publisher domain.Publisher) (*domain.Publisher, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Publisher, error)
	Update(ctx context.Context, conn *pgx.Conn, id int, publisher *domain.Publisher) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Publisher, error)
	Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Publisher, int, error)
	Exists(ctx context.Context, conn *pgx.Conn, id int) (bool, error)
	ExistsByName(ctx context.Context, conn *pgx.Conn, name string) (bool, error)
	ExistsByNameExcludeID(ctx context.Context, conn *pgx.Conn, name string, excludeID int) (bool, error)
	Count(ctx context.Context, conn *pgx.Conn) (int, error)
}

// BookCopyRepository определяет методы для работы с копиями книг
type BookCopyRepository interface {
	CreateCopy(ctx context.Context, conn *pgx.Conn, bookCopy *domain.BookCopy) error
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.BookCopy, error)
	GetCopiesByBookID(ctx context.Context, conn *pgx.Conn, bookID, limit, offset int) ([]domain.BookCopy, error)
	Update(ctx context.Context, conn *pgx.Conn, bookCopy domain.BookCopy) (*domain.BookCopy, error)
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	GetAvailable(ctx context.Context, conn *pgx.Conn, bookID int) ([]domain.BookCopy, error)
	UpdateStatus(ctx context.Context, conn *pgx.Conn, id int, status string) (*domain.BookCopy, error)
	CountAvailable(ctx context.Context, conn *pgx.Conn, bookID int) (int, error)
	HasActiveCopies(ctx context.Context, conn *pgx.Conn, bookID int) (bool, error)
	GetNextCopyNumber(ctx context.Context, conn *pgx.Conn, bookID int) (int, error)
	ExistsCopy(ctx context.Context, conn *pgx.Conn, copyID int) (bool, error)
	CountByBookID(ctx context.Context, conn *pgx.Conn, bookID int) (int, error)
	ClearReaderAndBorrowed(ctx context.Context, conn *pgx.Conn, id int) error
}

// ReaderRepository определяет методы для работы с читателями
type ReaderRepository interface {
	CreateReader(ctx context.Context, conn *pgx.Conn, reader *domain.Reader) (*domain.Reader, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.Reader, error)
	GetByEmail(ctx context.Context, conn *pgx.Conn, email string) (*domain.Reader, error)
	Update(ctx context.Context, conn *pgx.Conn, id int, reader domain.Reader) (*domain.Reader, error)
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error)
	GetActive(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error)
	GetDebtors(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error)
	BlockReader(ctx context.Context, conn *pgx.Conn, readerID int) error
	UnBlockReader(ctx context.Context, conn *pgx.Conn, readerID int) error
	IncrementBookCount(ctx context.Context, conn *pgx.Conn, readerID int) error
	DecrementBookCount(ctx context.Context, conn *pgx.Conn, readerID int) error
	UpdateStatus(ctx context.Context, conn *pgx.Conn, readerID int, reader domain.Reader) error
	Exists(ctx context.Context, conn *pgx.Conn, id int) (bool, error)
	ExistsEmail(ctx context.Context, conn *pgx.Conn, email string) (bool, error)
	ExistsPhone(ctx context.Context, conn *pgx.Conn, phone string) (bool, error)
	GetActiveBooksCount(ctx context.Context, conn *pgx.Conn, readerID int) (int, error)
	HasDebt(ctx context.Context, conn *pgx.Conn, readerID int) (bool, error)
}

// UserRepository определяет методы для работы с пользователями
type UserRepository interface {
	CreateUser(ctx context.Context, conn *pgx.Conn, user domain.User) error
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.User, error)
	GetByEmail(ctx context.Context, conn *pgx.Conn, email string) (domain.User, error)
	Update(ctx context.Context, conn *pgx.Conn, user domain.User) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	UpdatePassword(ctx context.Context, conn *pgx.Conn, id int, newPasswordHash string) error
	UpdateLastLogin(ctx context.Context, conn *pgx.Conn, id int) error
	DeleteByReaderID(ctx context.Context, conn *pgx.Conn, readerID int) error
}

// TransactionRepository определяет методы для работы с транзакциями
type TransactionRepository interface {
	CreateTransaction(ctx context.Context, conn *pgx.Conn, transaction domain.Transaction) error
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.Transaction, error)
	Update(ctx context.Context, conn *pgx.Conn, transaction domain.Transaction) error
	ListByReader(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Transaction, error)
	ListByBook(ctx context.Context, conn *pgx.Conn, copyID, limit, offset int) ([]domain.Transaction, error)
	GetActiveByReader(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Transaction, error)
	GetOverdue(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Transaction, error)
	GetByCopyID(ctx context.Context, conn *pgx.Conn, copyID, limit, offset int) ([]domain.Transaction, error)
	ReturnBook(ctx context.Context, conn *pgx.Conn, transactionID int, returnDate time.Time, fine float64) error
	BorrowBook(ctx context.Context, conn *pgx.Conn, copyID, readerID int, dueDate time.Time) error
	CountByReader(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Transaction, int, error)
	CountByBook(ctx context.Context, conn *pgx.Conn, bookID, limit, offset int) ([]domain.Transaction, int, error)
	IsTransactionActive(ctx context.Context, conn *pgx.Conn, transactionID int) (bool, error)
	HasReaderBorrowedBook(ctx context.Context, conn *pgx.Conn, readerID, bookID int) (bool, error)
}

// ReservationRepository определяет методы для работы с бронированиями
type ReservationRepository interface {
	CreateReservation(ctx context.Context, conn *pgx.Conn, reserv *domain.Reservation) (*domain.Reservation, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.Reservation, error)
	GetActiveByReader(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Reservation, int, error)
	GetActiveByCopy(ctx context.Context, conn *pgx.Conn, copyID, limit, offset int) ([]domain.Reservation, error)
	UpdateStatus(ctx context.Context, conn *pgx.Conn, id int, status string) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	GetExpired(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reservation, error)
	IsBookReservedByOther(ctx context.Context, conn *pgx.Conn, copyID, readerID int) (bool, error)
	HasActiveForCopy(ctx context.Context, conn *pgx.Conn, copyID int) (bool, error)
}

// ReviewRepository определяет методы для работы с отзывами
type ReviewRepository interface {
	CreateReview(ctx context.Context, conn *pgx.Conn, review domain.Review) (*domain.Review, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Review, error)
	Update(ctx context.Context, conn *pgx.Conn, review domain.Review) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	GetReviewsByBookID(ctx context.Context, conn *pgx.Conn, bookID, limit, offset int) ([]domain.Review, error)
	GetReviewsByReaderID(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Review, error)
	GetAverageRating(ctx context.Context, conn *pgx.Conn, bookID int) (float64, error)
	GetReviewCount(ctx context.Context, conn *pgx.Conn, bookID int) (int, error)
	Exists(ctx context.Context, conn *pgx.Conn, bookID, readerID int) (bool, error)
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Review, error)
	Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Review, int, error)
}

// SettingRepository определяет методы для работы с настройками
type SettingRepository interface {
	CreateSetting(ctx context.Context, conn *pgx.Conn, setting domain.Setting) error
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Setting, error)
	GetByKey(ctx context.Context, conn *pgx.Conn, key string) (domain.Setting, error)
	Update(ctx context.Context, conn *pgx.Conn, setting domain.Setting) error
	UpdateByKey(ctx context.Context, conn *pgx.Conn, key, value string) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	DeleteByKey(ctx context.Context, conn *pgx.Conn, key string) error
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Setting, error)
	Exists(ctx context.Context, conn *pgx.Conn, key string) (bool, error)
}

// AuditLogRepository определяет методы для работы с аудит-логами
type AuditLogRepository interface {
	CreateAuditLog(ctx context.Context, conn *pgx.Conn, log audit.AuditLog) error
}
