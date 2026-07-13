package repository

import (
	"context"
	"library/internal/domain"
	"library/internal/infrastructure/audit"
	"library/internal/infrastructure/settings"
	"time"

	"github.com/jackc/pgx/v5"
)

type BookRepository interface {
	CreateBook(ctx context.Context, conn *pgx.Conn, book domain.Book) (*domain.Book, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Book, error)
	GetByISBN(ctx context.Context, conn *pgx.Conn, isbn string) (domain.Book, error)
	Update(ctx context.Context, conn *pgx.Conn, id int, book domain.Book) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, error)
	Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Book, int, error)
	GetPopular(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Book, error)
	Exists(ctx context.Context, conn *pgx.Conn, id int) (bool, error)
	ExistsByISBN(ctx context.Context, conn *pgx.Conn, isbn string) (bool, error)
	Count(ctx context.Context, conn *pgx.Conn) (int, error)
	GetByPublisherID(ctx context.Context, conn *pgx.Conn, publisherID, limit, offset int) ([]domain.Book, error)
	CountByPublisherID(ctx context.Context, conn *pgx.Conn, publisherID int) (int, error)
	UpdateRating(ctx context.Context, conn *pgx.Conn, bookID int, avgRating float64) error
	UpdateRatingAndCount(ctx context.Context, conn *pgx.Conn, bookID int, reviewsCount int) error
}

type AuthorRepository interface {
	CreateAuthor(ctx context.Context, conn *pgx.Conn, author *domain.Author) (*domain.Author, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Author, error)
	Update(ctx context.Context, conn *pgx.Conn, author domain.Author) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Author, error)
	GetByBookID(ctx context.Context, conn *pgx.Conn, bookID int) ([]domain.Author, error)
	Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Author, int, error)
	Exists(ctx context.Context, conn *pgx.Conn, id int) (bool, error)
	CreateBookAuthor(ctx context.Context, conn *pgx.Conn, bookID, authorID int) error
	DeleteBookAuthorsByBookID(ctx context.Context, conn *pgx.Conn, bookID int) error
	ExistsByName(ctx context.Context, conn *pgx.Conn, firstName, lastName string) (bool, error)
	ExistsByNameExcludeID(ctx context.Context, conn *pgx.Conn, firstName, lastName string, excludeID int) (bool, error)
	CountAuthor(ctx context.Context, conn *pgx.Conn) (int, error)
	GetBooksByAuthorID(ctx context.Context, conn *pgx.Conn, authorID int) ([]domain.Book, error)
}

type GenreRepository interface {
	CreateGenre(ctx context.Context, conn *pgx.Conn, genre domain.Genre) (*domain.Genre, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.Genre, error)
	Update(ctx context.Context, conn *pgx.Conn, genre domain.Genre) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) (*domain.Genre, error)
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Genre, error)
	Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Genre, int, error)
	Exists(ctx context.Context, conn *pgx.Conn, id int) (bool, error)
	ExistsByNameGenre(ctx context.Context, conn *pgx.Conn, name string) (bool, error)
	GetSubGenres(ctx context.Context, conn *pgx.Conn, parentID int) ([]domain.Genre, error)
	GetRootGenres(ctx context.Context, conn *pgx.Conn) ([]domain.Genre, error)
	CreateBookGenre(ctx context.Context, conn *pgx.Conn, bookID, genreID int) error
	DeleteBookGenresByBookID(ctx context.Context, conn *pgx.Conn, bookID int) error
	ExistsByNameExcludeIDGenre(ctx context.Context, conn *pgx.Conn, name string, excludeID int) (bool, error)
	CountSubGenres(ctx context.Context, conn *pgx.Conn, genreID int) (int, error)
	CountBooksByGenreID(ctx context.Context, conn *pgx.Conn, genreID int) (int, error)
}

type PublisherRepository interface {
	Create(ctx context.Context, conn *pgx.Conn, publisher domain.Publisher) (*domain.Publisher, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Publisher, error)
	Update(ctx context.Context, conn *pgx.Conn, id int, publisher domain.Publisher) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Publisher, error)
	Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Publisher, int, error)
	Exists(ctx context.Context, conn *pgx.Conn, id int) (bool, error)
	ExistsByName(ctx context.Context, conn *pgx.Conn, name string) (bool, error)
	ExistsByNameExcludeID(ctx context.Context, conn *pgx.Conn, name string, excludeID int) (bool, error)
	Count(ctx context.Context, conn *pgx.Conn) (int, error)
	GetByPublisherID(ctx context.Context, conn *pgx.Conn, publisherID, limit, offset int) ([]domain.Book, error)
	CountByPublisherID(ctx context.Context, conn *pgx.Conn, publisherID int) (int, error)
}

type BookCopyRepository interface {
	CreateCopy(ctx context.Context, conn *pgx.Conn, copy *domain.BookCopy) error
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.BookCopy, error)
	GetCopiesByBookID(ctx context.Context, conn *pgx.Conn, bookID, limit, offset int) ([]domain.BookCopy, error)
	Update(ctx context.Context, conn *pgx.Conn, copy domain.BookCopy) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	GetAvailable(ctx context.Context, conn *pgx.Conn, bookID int) ([]domain.BookCopy, error)
	UpdateStatus(ctx context.Context, conn *pgx.Conn, id int, status string) error
	CountAvailable(ctx context.Context, conn *pgx.Conn, bookID int) (int, error)
	HasActiveCopies(ctx context.Context, conn *pgx.Conn, bookID int) (bool, error)
	GetNextCopyNumber(ctx context.Context, conn *pgx.Conn, bookID int) (int, error)
	ExistsCopy(ctx context.Context, conn *pgx.Conn, copyID int) (bool, error)
	CountByBookID(ctx context.Context, conn *pgx.Conn, bookID int) (int, error)
	ClearReaderAndBorrowed(ctx context.Context, conn *pgx.Conn, id int) error
}

type ReaderRepository interface {
	CreateReader(ctx context.Context, conn *pgx.Conn, reader *domain.Reader) (*domain.Reader, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.Reader, error)
	GetByEmail(ctx context.Context, conn *pgx.Conn, email string) (*domain.Reader, error)
	Update(ctx context.Context, conn *pgx.Conn, id int, reader domain.Reader) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error)
	GetActive(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error)
	GetDebtors(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reader, error)
	BlockReader(ctx context.Context, conn *pgx.Conn, readerId int) error
	UnBlockReader(ctx context.Context, conn *pgx.Conn, readerId int) error
	IncrementBookCount(ctx context.Context, conn *pgx.Conn, readerID int) error
	DecrementBookCount(ctx context.Context, conn *pgx.Conn, readerID int) error
	UpdateStatus(ctx context.Context, conn *pgx.Conn, readerID int, status string) error
	ExistsEmail(ctx context.Context, conn *pgx.Conn, email string) (bool, error)
	ExistsPhone(ctx context.Context, conn *pgx.Conn, phone string) (bool, error)
	Exists(ctx context.Context, conn *pgx.Conn, id int) (bool, error)
	GetActiveBooksCount(ctx context.Context, conn *pgx.Conn, readerID int) (int, error)
	HasDebt(ctx context.Context, conn *pgx.Conn, readerID int) (bool, error)
}

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
	CountByReader(ctx context.Context, conn *pgx.Conn, readerID int, limit, offset int) ([]domain.Transaction, int, error)
	CountByBook(ctx context.Context, conn *pgx.Conn, bookID int, limit, offset int) ([]domain.Transaction, int, error)
	IsTransactionActive(ctx context.Context, conn *pgx.Conn, transactionID int) (bool, error)
	HasReaderBorrowedBook(ctx context.Context, conn *pgx.Conn, readerID, bookID int) (bool, error)
}

type ReservationRepository interface {
	CreateReservation(ctx context.Context, conn *pgx.Conn, reservation *domain.Reservation) error
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (*domain.Reservation, error)
	GetActiveByReader(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Reservation, error)
	GetActiveByCopy(ctx context.Context, conn *pgx.Conn, copyID, limit, offset int) ([]domain.Reservation, error)
	UpdateStatus(ctx context.Context, conn *pgx.Conn, id int, newStatus string) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	GetExpired(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Reservation, error)
	IsBookReservedByOther(ctx context.Context, conn *pgx.Conn, copyID, readerID int) (bool, error)
	HasActiveForCopy(ctx context.Context, conn *pgx.Conn, copyID int) (bool, error)
}

type ReviewRepository interface {
	CreateReview(ctx context.Context, conn *pgx.Conn, review domain.Review) (*domain.Review, error)
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (domain.Review, error)
	Update(ctx context.Context, conn *pgx.Conn, review domain.Review) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]domain.Review, error)
	Search(ctx context.Context, conn *pgx.Conn, column, search string, limit, offset int) ([]domain.Review, int, error)
	GetReviewsByBookID(ctx context.Context, conn *pgx.Conn, bookID, limit, offset int) ([]domain.Review, error)
	GetReviewsByReaderID(ctx context.Context, conn *pgx.Conn, readerID, limit, offset int) ([]domain.Review, error)
	GetAverageRating(ctx context.Context, conn *pgx.Conn, bookID int) (float64, error)
	GetReviewCount(ctx context.Context, conn *pgx.Conn, bookID int) (int, error)
	Exists(ctx context.Context, conn *pgx.Conn, bookID, readerID int) (bool, error)
}

type AuditLogRepository interface {
	CreateAuditLog(ctx context.Context, conn *pgx.Conn, log audit.AuditLog) error
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (audit.AuditLog, error)
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]audit.AuditLog, error)
	GetByEntity(ctx context.Context, conn *pgx.Conn, entityType string, entityID, limit, offset int) ([]audit.AuditLog, error)
	GetByUser(ctx context.Context, conn *pgx.Conn, userID, limit, offset int) ([]audit.AuditLog, error)
	GetByAction(ctx context.Context, conn *pgx.Conn, action string, limit, offset int) ([]audit.AuditLog, error)
}

type SettingRepository interface {
	CreateSetting(ctx context.Context, conn *pgx.Conn, setting settings.Setting) error
	GetByID(ctx context.Context, conn *pgx.Conn, id int) (settings.Setting, error)
	GetByKey(ctx context.Context, conn *pgx.Conn, key string) (settings.Setting, error)
	Update(ctx context.Context, conn *pgx.Conn, setting settings.Setting) error
	UpdateByKey(ctx context.Context, conn *pgx.Conn, key, value string) error
	Delete(ctx context.Context, conn *pgx.Conn, id int) error
	DeleteByKey(ctx context.Context, conn *pgx.Conn, key string) error
	List(ctx context.Context, conn *pgx.Conn, limit, offset int) ([]settings.Setting, error)
	Exists(ctx context.Context, conn *pgx.Conn, key string) (bool, error)
}
