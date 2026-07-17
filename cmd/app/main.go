package main

import (
	"context"
	"library/internal/infrastructure/audit"
	"log"
	"os"

	"library/internal/handlers"
	"library/internal/server"
	"library/internal/service"
	"library/pkg/database"
	"library/pkg/logger"

	"library/internal/repository/postgres"

	"go.uber.org/zap"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	zapLogger, closeLogger, err := logger.NewLogger(logLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer closeLogger()

	zapLogger.Info("starting library service...")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:dani123l@localhost:5432/library?sslmode=disable"
	}

	conn, err := database.Connect(databaseURL)
	if err != nil {
		zapLogger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer conn.Close(context.Background())

	zapLogger.Info("database connected successfully")

	authorRepo := &postgres.AuthorRepository{}
	bookRepo := &postgres.BookRepository{}
	copyRepo := &postgres.BookCopyRepository{}
	genreRepo := &postgres.GenreRepository{}
	publisherRepo := &postgres.PublisherRepository{}
	readerRepo := &postgres.ReaderRepository{}
	userRepo := &postgres.UserRepository{}
	transactionRepo := &postgres.TransactionRepository{}
	reservationRepo := &postgres.ReservationRepository{}
	reviewRepo := &postgres.ReviewRepository{}
	settingRepo := &postgres.SettingRepository{}
	auditRepo := &audit.AuditLogRepository{}

	authorService := service.NewAuthorService(authorRepo, bookRepo, auditRepo, zapLogger)
	bookService := service.NewBookService(bookRepo, copyRepo, authorRepo, genreRepo, publisherRepo, auditRepo, zapLogger)
	copyService := service.NewCopyService(copyRepo, bookRepo, readerRepo, reservationRepo, zapLogger)
	genreService := service.NewGenreService(genreRepo, bookRepo, zapLogger)
	publisherService := service.NewPublisherService(publisherRepo, bookRepo, zapLogger)
	readerService := service.NewReaderService(readerRepo, userRepo, transactionRepo, zapLogger)
	reservationService := service.NewReservation(reservationRepo, copyRepo, readerRepo, bookRepo, transactionRepo, settingRepo, zapLogger)
	reviewService := service.NewReview(reviewRepo, bookRepo, readerRepo, transactionRepo, zapLogger)
	transactionService := service.NewTransactionService(transactionRepo, copyRepo, readerRepo, bookRepo, settingRepo, reservationRepo, zapLogger)
	settingService := service.NewSettingService(settingRepo, auditRepo, zapLogger)

	authorHandler := handlers.NewHTTPHandlersAuthor(authorService)
	copyHandler := handlers.NewHTTPHandlersCopy(copyService)
	bookHandler := handlers.NewBookHandler(bookService)
	genreHandler := handlers.NewGenreHandlers(genreService)
	publisherHandler := handlers.NewPublisherHandler(publisherService)
	readerHandler := handlers.NewReaderHandlers(readerService)
	reservationHandler := handlers.NewReservationHandler(reservationService, copyService, bookService, readerService)
	reviewHandler := handlers.NewReviewHandlers(reviewService, bookService, readerService)
	transactionHandler := handlers.NewTransactionHandler(transactionService, readerService, copyService, bookService)
	settingHandler := handlers.NewSettingsHandler(settingService)

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
		settingHandler,
	)

	zapLogger.Info("starting HTTP server on :9091")
	if err := httpServer.Start(); err != nil {
		zapLogger.Fatal("server failed", zap.Error(err))
	}
}
