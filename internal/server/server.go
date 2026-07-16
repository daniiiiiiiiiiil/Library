package server

import (
	"errors"
	"library/internal/handlers"
	"library/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type HTTPServer struct {
	httpHandlerAuthor      *handlers.HTTPHandlersAuthor
	httpHandlerCopy        *handlers.CopyHandler
	httpHandlerBook        *handlers.BookHandler
	httpHandlerGenre       *handlers.HTTPHandlersGenre
	httpHandlerPublisher   *handlers.HTTPHandlersPublisher
	httpHandlerReader      *handlers.ReaderHandlers
	httpHandlerReservation *handlers.HTTPHandlersReservation
	httpHandlerReview      *handlers.HTTPHandlersReview
	httpHandlerTransaction *handlers.HTTPHandlersTransaction
	conn                   *pgx.Conn
}

func NewHTTPServer(
	conn *pgx.Conn,
	httpHandlerAuthor *handlers.HTTPHandlersAuthor,
	httpHandlerCopy *handlers.CopyHandler,
	httpHandlerBook *handlers.BookHandler,
	httpHandlerGenre *handlers.HTTPHandlersGenre,
	httpHandlerPublisher *handlers.HTTPHandlersPublisher,
	httpHandlerReader *handlers.ReaderHandlers,
	httpHandlerReservation *handlers.HTTPHandlersReservation,
	httpHandlerReview *handlers.HTTPHandlersReview,
	httpHandlerTransaction *handlers.HTTPHandlersTransaction) *HTTPServer {
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
	}
}

func (server *HTTPServer) Start() error {
	router := mux.NewRouter()
	router.Use(middleware.DBConnection(server.conn))
	//книги
	router.Path("/api/v1/books").Methods("POST").HandlerFunc(server.httpHandlerBook.CreateBook)
	router.Path("/api/v1/books").Methods("GET").HandlerFunc(server.httpHandlerBook.ListBooks)
	router.Path("/api/v1/books/{id}").HandlerFunc(server.httpHandlerBook.GetBook)
	router.Path("/api/v1/books/isbn/{isbn}").Methods("GET").HandlerFunc(server.httpHandlerBook.GetBookByISBN)
	router.Path("/api/v1/books/{id}").Methods("PUT").HandlerFunc(server.httpHandlerBook.UpdateBook)
	router.Path("/api/v1/books/{id}").Methods("DELETE").HandlerFunc(server.httpHandlerBook.DeleteBook)
	router.Path("/api/v1/books/search").Methods("GET").HandlerFunc(server.httpHandlerBook.SearchBooks)
	router.Path("/api/v1/books/popular").Methods("GET").HandlerFunc(server.httpHandlerBook.GetPopularBooks)
	router.Path("/api/v1/books/{id}/copies").Methods("POST").HandlerFunc(server.httpHandlerBook.AddCopy)

	if err := http.ListenAndServe(":9091", router); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
	}
	return nil
}
