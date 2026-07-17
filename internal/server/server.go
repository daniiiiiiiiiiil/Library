package server

import (
	"encoding/json"
	"errors"
	"library/internal/handlers"
	"library/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type HTTPServer struct {
	conn                   *pgx.Conn
	httpHandlerAuthor      *handlers.AuthorHandler
	httpHandlerCopy        *handlers.CopyHandler
	httpHandlerBook        *handlers.BookHandler
	httpHandlerGenre       *handlers.GenreHandlers
	httpHandlerPublisher   *handlers.PublisherHandler
	httpHandlerReader      *handlers.ReaderHandlers
	httpHandlerReservation *handlers.ReservationHandler
	httpHandlerReview      *handlers.ReviewHandlers
	httpHandlerTransaction *handlers.TransactionHandler
	httpHandlerSettings    *handlers.SettingsHandler
}

func NewHTTPServer(
	conn *pgx.Conn,
	httpHandlerAuthor *handlers.AuthorHandler,
	httpHandlerCopy *handlers.CopyHandler,
	httpHandlerBook *handlers.BookHandler,
	httpHandlerGenre *handlers.GenreHandlers,
	httpHandlerPublisher *handlers.PublisherHandler,
	httpHandlerReader *handlers.ReaderHandlers,
	httpHandlerReservation *handlers.ReservationHandler,
	httpHandlerReview *handlers.ReviewHandlers,
	httpHandlerTransaction *handlers.TransactionHandler,
	httpHandlerSettings *handlers.SettingsHandler,
) *HTTPServer {
	return &HTTPServer{
		conn:                   conn,
		httpHandlerAuthor:      httpHandlerAuthor,
		httpHandlerCopy:        httpHandlerCopy,
		httpHandlerBook:        httpHandlerBook,
		httpHandlerGenre:       httpHandlerGenre,
		httpHandlerPublisher:   httpHandlerPublisher,
		httpHandlerReader:      httpHandlerReader,
		httpHandlerReservation: httpHandlerReservation,
		httpHandlerReview:      httpHandlerReview,
		httpHandlerTransaction: httpHandlerTransaction,
		httpHandlerSettings:    httpHandlerSettings,
	}
}

func (server *HTTPServer) Start() error {
	router := mux.NewRouter()
	router.Use(middleware.DBConnection(server.conn))

	// HEALTH CHECK
	router.Path("/api/v1/health").Methods("GET").HandlerFunc(server.healthCheck)

	// AUTHORS
	router.Path("/api/v1/authors/search").Methods("GET").HandlerFunc(server.httpHandlerAuthor.GetAuthorSearch)
	router.Path("/api/v1/authors/book/{id}").Methods("GET").HandlerFunc(server.httpHandlerAuthor.AuthorsByBook)
	router.Path("/api/v1/authors").Methods("POST").HandlerFunc(server.httpHandlerAuthor.PostCreateAuthor)
	router.Path("/api/v1/authors").Methods("GET").HandlerFunc(server.httpHandlerAuthor.GetAuthors)
	router.Path("/api/v1/authors/{id}").Methods("GET").HandlerFunc(server.httpHandlerAuthor.GetAuthor)
	router.Path("/api/v1/authors/{id}").Methods("PUT").HandlerFunc(server.httpHandlerAuthor.UpdateAuthor)
	router.Path("/api/v1/authors/{id}").Methods("DELETE").HandlerFunc(server.httpHandlerAuthor.DeleteAuthor)

	// BOOKS
	router.Path("/api/v1/books/search").Methods("GET").HandlerFunc(server.httpHandlerBook.SearchBooks)
	router.Path("/api/v1/books/popular").Methods("GET").HandlerFunc(server.httpHandlerBook.GetPopularBooks)
	router.Path("/api/v1/books/isbn/{isbn}").Methods("GET").HandlerFunc(server.httpHandlerBook.GetBookByISBN)
	router.Path("/api/v1/books").Methods("POST").HandlerFunc(server.httpHandlerBook.CreateBook)
	router.Path("/api/v1/books").Methods("GET").HandlerFunc(server.httpHandlerBook.ListBooks)
	router.Path("/api/v1/books/{id}/copies").Methods("POST").HandlerFunc(server.httpHandlerBook.AddCopy)
	router.Path("/api/v1/books/{id}").Methods("GET").HandlerFunc(server.httpHandlerBook.GetBook)
	router.Path("/api/v1/books/{id}").Methods("PUT").HandlerFunc(server.httpHandlerBook.UpdateBook)
	router.Path("/api/v1/books/{id}").Methods("DELETE").HandlerFunc(server.httpHandlerBook.DeleteBook)

	//COPIES
	router.Path("/api/v1/copies/book/{bookId}/available").Methods("GET").HandlerFunc(server.httpHandlerCopy.GetCopiesAvaliables)
	router.Path("/api/v1/copies/book/{bookId}").Methods("GET").HandlerFunc(server.httpHandlerCopy.GetCopiesByBook)
	router.Path("/api/v1/copies/{CopyID}/status").Methods("PATCH").HandlerFunc(server.httpHandlerCopy.PathCopyStatus)
	router.Path("/api/v1/copies/{CopyID}").Methods("GET").HandlerFunc(server.httpHandlerCopy.GetCopy)
	router.Path("/api/v1/copies/{CopyID}").Methods("PUT").HandlerFunc(server.httpHandlerCopy.PUTCopies)
	router.Path("/api/v1/copies/{CopyID}").Methods("DELETE").HandlerFunc(server.httpHandlerCopy.DeleteCopies)
	router.Path("/api/v1/copies/book/{CopyID}/count").Methods("GET").HandlerFunc(server.httpHandlerCopy.GETCountAvailableCopy)

	// GENRES
	router.Path("/api/v1/genres/search").Methods("GET").HandlerFunc(server.httpHandlerGenre.GenreSearch)
	router.Path("/api/v1/genres/hierarchy").Methods("GET").HandlerFunc(server.httpHandlerGenre.GetGenreHierarchy)
	router.Path("/api/v1/genres/sub/{parent_id}").Methods("GET").HandlerFunc(server.httpHandlerGenre.GetSubgenres)
	router.Path("/api/v1/genres").Methods("POST").HandlerFunc(server.httpHandlerGenre.CreateGenre)
	router.Path("/api/v1/genres").Methods("GET").HandlerFunc(server.httpHandlerGenre.GetGenres)
	router.Path("/api/v1/genres/{genreID}").Methods("GET").HandlerFunc(server.httpHandlerGenre.GetGenre)
	router.Path("/api/v1/genres/{genreID}").Methods("PUT").HandlerFunc(server.httpHandlerGenre.UpdateGenre)
	router.Path("/api/v1/genres/{genreID}").Methods("DELETE").HandlerFunc(server.httpHandlerGenre.DeleteGenre)

	// PUBLISHERS
	router.Path("/api/v1/publishers/search").Methods("GET").HandlerFunc(server.httpHandlerPublisher.SearchPublisher)
	router.Path("/api/v1/publishers/{publisherId}/books").Methods("GET").HandlerFunc(server.httpHandlerPublisher.GetBooksPublisher)
	router.Path("/api/v1/publishers").Methods("POST").HandlerFunc(server.httpHandlerPublisher.CreatePublisher)
	router.Path("/api/v1/publishers").Methods("GET").HandlerFunc(server.httpHandlerPublisher.GetPublishers)
	router.Path("/api/v1/publishers/{publisherId}").Methods("GET").HandlerFunc(server.httpHandlerPublisher.GetPublisher)
	router.Path("/api/v1/publishers/{publisherId}").Methods("PUT").HandlerFunc(server.httpHandlerPublisher.UpdatePublisher)
	router.Path("/api/v1/publishers/{publisherId}").Methods("DELETE").HandlerFunc(server.httpHandlerPublisher.DeletePublisher)

	// READERS
	router.Path("/api/v1/readers/active").Methods("GET").HandlerFunc(server.httpHandlerReader.GetActiveReaders)
	router.Path("/api/v1/readers/debtors").Methods("GET").HandlerFunc(server.httpHandlerReader.GetDebtors)
	router.Path("/api/v1/readers/email/{email}").Methods("GET").HandlerFunc(server.httpHandlerReader.GetReaderEmail)
	router.Path("/api/v1/readers").Methods("POST").HandlerFunc(server.httpHandlerReader.CreateReader)
	router.Path("/api/v1/readers").Methods("GET").HandlerFunc(server.httpHandlerReader.GetReader)
	router.Path("/api/v1/readers/{id}/block").Methods("POST").HandlerFunc(server.httpHandlerReader.BlockReader)
	router.Path("/api/v1/readers/{id}/unblock").Methods("POST").HandlerFunc(server.httpHandlerReader.UnblockReader)
	router.Path("/api/v1/readers/{id}/history").Methods("GET").HandlerFunc(server.httpHandlerReader.GetReaderHistory)
	router.Path("/api/v1/readers/{id}").Methods("GET").HandlerFunc(server.httpHandlerReader.GetReaderID)
	router.Path("/api/v1/readers/{id}").Methods("PUT").HandlerFunc(server.httpHandlerReader.UpdateReader)
	router.Path("/api/v1/readers/{id}").Methods("DELETE").HandlerFunc(server.httpHandlerReader.DeleteReader)

	// RESERVATIONS
	router.Path("/api/v1/reservations/reader/{reader_id}/copy/{copy_id}").Methods("POST").HandlerFunc(server.httpHandlerReservation.CreateReservationBook)
	router.Path("/api/v1/reservations/reader/{reader_id}").Methods("GET").HandlerFunc(server.httpHandlerReservation.GetActiveReservedBookReader)
	router.Path("/api/v1/reservations/copy/{copyId}").Methods("GET").HandlerFunc(server.httpHandlerReservation.GetActiveReservationsByCopy)
	router.Path("/api/v1/reservations/expired/process").Methods("POST").HandlerFunc(server.httpHandlerReservation.ProcessExpiredReservations)
	router.Path("/api/v1/reservations/{reservation_id}").Methods("DELETE").HandlerFunc(server.httpHandlerReservation.DeleteReservation)
	router.Path("/api/v1/reservations/{reservation_id}").Methods("GET").HandlerFunc(server.httpHandlerReservation.GetReservationByID)

	// REVIEWS
	router.Path("/api/v1/reviews/reader/{reader_id}").Methods("POST").HandlerFunc(server.httpHandlerReview.CreateReview)
	router.Path("/api/v1/reviews/reader/{reader_id}/list").Methods("GET").HandlerFunc(server.httpHandlerReview.GetReviewReader)
	router.Path("/api/v1/reviews/book/{book_id}").Methods("GET").HandlerFunc(server.httpHandlerReview.GetReviewBook)
	router.Path("/api/v1/reviews/book/{bookId}/rating").Methods("GET").HandlerFunc(server.httpHandlerReview.GetBookRating)
	router.Path("/api/v1/reviews/{review_id}").Methods("GET").HandlerFunc(server.httpHandlerReview.GetReview)
	router.Path("/api/v1/reviews/{id}").Methods("PUT").HandlerFunc(server.httpHandlerReview.UpdateReview)
	router.Path("/api/v1/reviews/{id}").Methods("DELETE").HandlerFunc(server.httpHandlerReview.DeleteReview)

	// TRANSACTIONS
	router.Path("/api/v1/transactions/borrow").Methods("POST").HandlerFunc(server.httpHandlerTransaction.BorrowBook)
	router.Path("/api/v1/transactions/return").Methods("POST").HandlerFunc(server.httpHandlerTransaction.ReturnBook)
	router.Path("/api/v1/transactions/overdue").Methods("GET").HandlerFunc(server.httpHandlerTransaction.GetOverdueTransactions)
	router.Path("/api/v1/transactions/overdue/process").Methods("POST").HandlerFunc(server.httpHandlerTransaction.ProcessOverdueTransactions)
	router.Path("/api/v1/transactions/reader/{reader_id}").Methods("GET").HandlerFunc(server.httpHandlerTransaction.GetTransactionByReader)
	router.Path("/api/v1/transactions/reader/{readerId}/active").Methods("GET").HandlerFunc(server.httpHandlerTransaction.GetActiveReaderTransactions)
	router.Path("/api/v1/transactions/book/{bookId}").Methods("GET").HandlerFunc(server.httpHandlerTransaction.GetBookTransactions)
	router.Path("/api/v1/transactions/{transaction_id}").Methods("GET").HandlerFunc(server.httpHandlerTransaction.GetTransactions)

	// SETTINGS
	router.Path("/api/v1/settings/key/{key}").Methods("GET").HandlerFunc(server.httpHandlerSettings.GetSettingByKey)
	router.Path("/api/v1/settings/key/{key}").Methods("PUT").HandlerFunc(server.httpHandlerSettings.UpdateSettingByKey)
	router.Path("/api/v1/settings/key/{key}").Methods("DELETE").HandlerFunc(server.httpHandlerSettings.DeleteSettingByKey)
	router.Path("/api/v1/settings").Methods("GET").HandlerFunc(server.httpHandlerSettings.GetSettings)
	router.Path("/api/v1/settings").Methods("POST").HandlerFunc(server.httpHandlerSettings.CreateSetting)
	router.Path("/api/v1/settings/{id}").Methods("GET").HandlerFunc(server.httpHandlerSettings.GetSettingByID)
	router.Path("/api/v1/settings/{id}").Methods("PUT").HandlerFunc(server.httpHandlerSettings.UpdateSetting)
	router.Path("/api/v1/settings/{id}").Methods("DELETE").HandlerFunc(server.httpHandlerSettings.DeleteSetting)

	port := ":9091"
	if err := http.ListenAndServe(port, router); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
	return nil
}

func (server *HTTPServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"service": "library-api",
		"version": "1.0.0",
	})
}
