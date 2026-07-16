package main

import (
	"context"
	"library/internal/handlers"
	"library/internal/infrastructure/audit"
	"library/internal/repository/postgres"
	"library/internal/server"
	"library/internal/service"
	"library/pkg/database"
	"library/pkg/logger"
	"log"

	"go.uber.org/zap"
)

func main() {
	logg, closeLog, err := logger.NewLogger("DEBUG")
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer closeLog()

	conn, err := database.Connect("postgres://postgres:dani123l@localhost:5432/library?sslmode=disable")
	if err != nil {
		logg.Fatal("Failed to connect to database:", zap.Error(err))
	}
	defer conn.Close(context.Background())

	ctx := context.Background()
	_ = ctx

	bookRepo := &postgres.BookRepository{}
	authorRepo := &postgres.AuthorRepository{}
	genreRepo := &postgres.GenreRepository{}
	publisherRepo := &postgres.PublisherRepository{}
	copyRepo := &postgres.BookCopyRepository{}
	readerRepo := &postgres.ReaderRepository{}
	userRepo := &postgres.UserRepository{}
	txRepo := &postgres.TransactionRepository{}
	reservationRepo := &postgres.ReservationRepository{}
	reviewRepo := &postgres.ReviewRepository{}
	auditRepo := &audit.AuditLogRepository{}
	settingRepo := &postgres.SettingRepository{}

	bookService := service.NewBookService(
		bookRepo,
		copyRepo,
		authorRepo,
		genreRepo,
		publisherRepo,
		auditRepo,
		logg,
	)
	authorService := service.NewAuthorService(
		authorRepo,
		bookRepo,
		auditRepo,
		logg,
	)
	genreService := service.NewGenreService(
		genreRepo,
		bookRepo,
		logg,
	)
	publisherService := service.NewPublisherService(
		publisherRepo,
		bookRepo,
		logg,
	)
	copyService := service.NewCopyService(
		copyRepo,
		bookRepo,
		readerRepo,
		reservationRepo,
		logg,
	)
	readerService := service.NewReaderService(
		readerRepo,
		userRepo,
		txRepo,
		logg,
	)
	transactionService := service.NewTransactionService(
		txRepo,
		copyRepo,
		readerRepo,
		bookRepo,
		settingRepo,
		reservationRepo,
		logg,
	)
	reservationService := service.NewReservation(
		reservationRepo,
		copyRepo,
		readerRepo,
		bookRepo,
		txRepo,
		settingRepo,
		logg,
	)
	reviewService := service.NewReview(
		reviewRepo,
		bookRepo,
		readerRepo,
		txRepo,
		logg,
	)
	settingService := service.NewSettingService(
		settingRepo,
		auditRepo,
		logg,
	)

	bookHandler := handlers.NewBookHandler(bookService)
	authorHandler := handlers.NewHTTPHandlersAuthor(authorService)
	copyHandler := handlers.NewHTTPHandlersCopy(copyService)
	genreHandler := handlers.NewGenreHandlers(genreService)
	publisherHandler := handlers.NewPublisherHandler(publisherService)
	readerHandler := handlers.NewReaderHandlers(readerService)
	reservationHandler := handlers.NewReservationHandler(reservationService, copyService, bookService, readerService)
	reviewHandler := handlers.NewReviewHandlers(reviewService, bookService, readerService)
	transactionHandler := handlers.NewHTTPHandlersTransaction(transactionService)

	httpServer := server.NewHTTPServer(
		conn,
		authorHandler,
		copyHandler,
		bookHandler,
		genreHandler,
		publisherHandler,
		readerHandler,
		reservationHandler,
		reviewHandler,
		transactionHandler,
	)

	logg.Info("Starting HTTP server on :9091")
	if err := httpServer.Start(); err != nil {
		logg.Fatal("Failed to start HTTP server:", zap.Error(err))
	}
}
