package main

import (
	"context"
	"library/internal/domain"
	"library/internal/infrastructure/audit"
	"library/internal/repository/postgres"
	"library/internal/service"
	"library/pkg/database"
	"library/pkg/logger"
	"log"
	"time"

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

	author, err := authorService.CreateAuthor(ctx, conn, &domain.Author{
		First_name: "Лев",
		Last_name:  "Толстой",
		Biography:  "Великий русский писатель",
		Birthday:   time.Date(1828, 9, 9, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		logg.Warn("CreateAuthor error (may already exist)", zap.Error(err))
		authors, _, _ := authorService.ListAuthors(ctx, conn, 10, 0)
		for _, a := range authors {
			if a.First_name == "Лев" && a.Last_name == "Толстой" {
				author = &a
				logg.Info("Found existing author", zap.Int("id", author.ID), zap.String("name", author.First_name+" "+author.Last_name))
				break
			}
		}
		if author == nil {
			logg.Fatal("Cannot get or create author")
		}
	} else {
		logg.Info("Author created successfully", zap.Int("id", author.ID), zap.String("name", author.First_name+" "+author.Last_name))
	}

	publisher, err := publisherService.CreatePublisher(ctx, conn, &domain.Publisher{
		Name:    "АСТ",
		Address: "г. Москва",
		Phone:   "+7 (495) 123-45-67",
	})
	if err != nil {
		logg.Warn("CreatePublisher error (may already exist)", zap.Error(err))
		publishers, _, _ := publisherService.ListPublishers(ctx, conn, 10, 0)
		for _, p := range publishers {
			if p.Name == "АСТ" {
				publisher = &p
				logg.Info("Found existing publisher", zap.Int("id", publisher.ID), zap.String("name", publisher.Name))
				break
			}
		}
		if publisher == nil {
			logg.Fatal("Cannot get or create publisher")
		}
	} else {
		logg.Info("Publisher created successfully", zap.Int("id", publisher.ID), zap.String("name", publisher.Name))
	}

	genre, err := genreService.CreateGenre(ctx, conn, domain.Genre{
		Name: "Роман",
	})
	if err != nil {
		logg.Warn("CreateGenre error (may already exist)", zap.Error(err))
		genres, _ := genreService.ListGenres(ctx, conn, 10, 0)
		for _, g := range genres {
			if g.Name == "Роман" {
				genre = &g
				logg.Info("Found existing genre", zap.Int("id", genre.ID), zap.String("name", genre.Name))
				break
			}
		}
		if genre == nil {
			logg.Fatal("Cannot get or create genre")
		}
	} else {
		logg.Info("Genre created successfully", zap.Int("id", genre.ID), zap.String("name", genre.Name))
	}

	logg.Info("Creating book",
		zap.Int("author_id", author.ID),
		zap.Int("publisher_id", publisher.ID),
		zap.Int("genre_id", genre.ID),
	)

	book, err := bookService.CreateBook(
		ctx,
		conn,
		&domain.Book{
			Title:         "Война и мир",
			ISBN:          "978-5-17-118421-5",
			Year:          1869,
			PublisherID:   publisher.ID,
			Description:   "Роман-эпопея Льва Толстого",
			CoverImageURL: "https://example.com/cover.jpg",
		},
		[]int{author.ID},
		[]int{genre.ID},
	)
	if err != nil {
		logg.Error("CreateBook error", zap.Error(err))
		book, err = bookService.GetBookByISBN(ctx, conn, "978-5-17-118421-5")
		if err != nil {
			logg.Error("Cannot get or create book", zap.Error(err))
		} else {
			logg.Info("Found existing book", zap.Int("id", book.ID), zap.String("title", book.Title))
		}
	} else {
		logg.Info("Book created successfully", zap.Int("id", book.ID), zap.String("title", book.Title))
	}

	if book != nil {
		err = bookService.AddCopyToBook(ctx, conn, book.ID, "good")
		if err != nil {
			logg.Debug("AddCopyToBook (good) error (may already exist)", zap.Int("book_id", book.ID), zap.Error(err))
		} else {
			logg.Info("AddCopyToBook success", zap.Int("book_id", book.ID), zap.String("condition", "good"))
		}

		err = bookService.AddCopyToBook(ctx, conn, book.ID, "excellent")
		if err != nil {
			logg.Debug("AddCopyToBook (excellent) error (may already exist)", zap.Int("book_id", book.ID), zap.Error(err))
		} else {
			logg.Info("AddCopyToBook success", zap.Int("book_id", book.ID), zap.String("condition", "excellent"))
		}
	}

	if book != nil {
		gotBook, err := bookService.GetBook(ctx, conn, book.ID)
		if err != nil {
			logg.Error("GetBook error", zap.Int("book_id", book.ID), zap.Error(err))
		} else {
			logg.Info("GetBook success", zap.Int("id", gotBook.ID), zap.String("title", gotBook.Title))
		}

		gotBookByISBN, err := bookService.GetBookByISBN(ctx, conn, "978-5-17-118421-5")
		if err != nil {
			logg.Error("GetBookByISBN error", zap.Error(err))
		} else {
			logg.Info("GetBookByISBN success", zap.Int("id", gotBookByISBN.ID), zap.String("title", gotBookByISBN.Title))
		}
	}

	books, total, err := bookService.ListBooks(ctx, conn, 10, 0)
	if err != nil {
		logg.Error("ListBooks error", zap.Error(err))
	} else {
		logg.Info("ListBooks success", zap.Int("returned", len(books)), zap.Int("total", total))
	}

	searchResults, count, err := bookService.SearchBooks(ctx, conn, "title", "Война", 10, 0)
	if err != nil {
		logg.Error("SearchBooks error", zap.Error(err))
	} else {
		logg.Info("SearchBooks success", zap.Int("found", count))
		for _, b := range searchResults {
			logg.Debug("SearchBooks result", zap.String("title", b.Title), zap.String("isbn", b.ISBN))
		}
	}

	popularBooks, err := bookService.GetPopularBooks(ctx, conn, 10, 0)
	if err != nil {
		logg.Error("GetPopularBooks error", zap.Error(err))
	} else {
		logg.Info("GetPopularBooks success", zap.Int("count", len(popularBooks)))
	}

	reader, err := readerService.CreateReader(ctx, conn, &domain.Reader{
		Name:     "Иван Петров",
		Phone:    "+7 (999) 123-45-67",
		Email:    "ivan@example.com",
		MaxBooks: 5,
	}, "secure_password123")
	if err != nil {
		logg.Warn("CreateReader error (may already exist)", zap.Error(err))
		reader, err = readerService.GetByEmail(ctx, conn, "ivan@example.com")
		if err != nil {
			logg.Error("Cannot get or create reader", zap.Error(err))
		} else {
			logg.Info("Found existing reader", zap.Int("id", reader.Id), zap.String("name", reader.Name))
		}
	} else {
		logg.Info("Reader created successfully", zap.Int("id", reader.Id), zap.String("name", reader.Name))
	}

	if book != nil && reader != nil {
		copies, err := copyService.GetAvailableCopies(ctx, conn, book.ID)
		if err != nil {
			logg.Error("GetAvailableCopies error", zap.Int("book_id", book.ID), zap.Error(err))
		} else if len(copies) > 0 {
			transaction, err := transactionService.BorrowBook(
				ctx,
				conn,
				copies[0].ID,
				reader.Id,
				time.Now().AddDate(0, 0, 14),
			)
			if err != nil {
				logg.Debug("BorrowBook error (may already borrowed)", zap.Error(err))
			} else {
				logg.Info("BorrowBook success", zap.Int("transaction_id", transaction.ID), zap.Int("copy_id", copies[0].ID))
			}
		} else {
			logg.Warn("No available copies for borrowing", zap.Int("book_id", book.ID))
		}
	}

	if reader != nil {
		gotReader, err := readerService.GetReader(ctx, conn, reader.Id)
		if err != nil {
			logg.Error("GetReader error", zap.Int("reader_id", reader.Id), zap.Error(err))
		} else {
			logg.Info("GetReader success", zap.String("name", gotReader.Name), zap.Int("books", gotReader.BooksCount), zap.Int("max_books", gotReader.MaxBooks))
		}
	}

	if book != nil && reader != nil {
		copies, err := copyService.GetAvailableCopies(ctx, conn, book.ID)
		if err != nil {
			logg.Error("GetAvailableCopies for reservation error", zap.Int("book_id", book.ID), zap.Error(err))
		} else if len(copies) > 0 {
			err = reservationService.ReserveBook(ctx, conn, copies[0].ID, reader.Id)
			if err != nil {
				logg.Debug("ReserveBook error (may already reserved)", zap.Error(err))
			} else {
				logg.Info("ReserveBook success", zap.Int("copy_id", copies[0].ID), zap.Int("reader_id", reader.Id))
			}
		} else {
			logg.Warn("No available copies for reservation", zap.Int("book_id", book.ID))
		}
	}

	if book != nil && reader != nil {
		review, err := reviewService.CreateReview(ctx, conn, domain.Review{
			BookID:   book.ID,
			ReaderID: reader.Id,
			Rating:   4.5,
			Comment:  "Отличная книга!",
		})
		if err != nil {
			logg.Debug("CreateReview error (may already exist)", zap.Error(err))
		} else {
			logg.Info("CreateReview success", zap.Int("book_id", review.BookID), zap.Float64("rating", review.Rating))
		}
	}

	if book != nil {
		avgRating, reviewCount, err := reviewService.GetBookRating(ctx, conn, book.ID)
		if err != nil {
			logg.Error("GetBookRating error", zap.Int("book_id", book.ID), zap.Error(err))
		} else {
			logg.Info("GetBookRating success", zap.Int("book_id", book.ID), zap.Float64("avg_rating", avgRating), zap.Int("reviews", reviewCount))
		}
	}

	setting, err := settingService.CreateSetting(ctx, conn, &domain.Setting{
		Key:         "fine_rate_per_day",
		Value:       "1.50",
		Description: "Штраф за просрочку",
	})
	if err != nil {
		logg.Debug("CreateSetting error (may already exist)", zap.Error(err))
	} else {
		logg.Info("CreateSetting success", zap.String("key", setting.Key), zap.String("value", setting.Value))
	}

	settingValue, err := settingService.GetSettingValue(ctx, conn, "fine_rate_per_day")
	if err != nil {
		logg.Error("GetSettingValue error", zap.Error(err))
	} else {
		logg.Info("GetSettingValue success", zap.String("value", settingValue))
	}

	fineRate, err := settingService.GetSettingFloat(ctx, conn, "fine_rate_per_day")
	if err != nil {
		logg.Error("GetSettingFloat error", zap.Error(err))
	} else {
		logg.Info("GetSettingFloat success", zap.Float64("fine_rate", fineRate))
	}

	authors, totalAuthors, err := authorService.ListAuthors(ctx, conn, 10, 0)
	if err != nil {
		logg.Error("ListAuthors error", zap.Error(err))
	} else {
		logg.Info("ListAuthors success", zap.Int("returned", len(authors)), zap.Int("total", totalAuthors))
		for _, a := range authors {
			logg.Debug("ListAuthors result", zap.Int("id", a.ID), zap.String("name", a.First_name+" "+a.Last_name))
		}
	}

	if book != nil {
		updatedBook, err := bookService.UpdateBook(
			ctx,
			conn,
			book.ID,
			map[string]interface{}{
				"title":       "Война и мир (обновлённое издание)",
				"description": "Обновлённое описание",
			},
			[]int{author.ID},
			[]int{genre.ID},
		)
		if err != nil {
			logg.Error("UpdateBook error", zap.Int("book_id", book.ID), zap.Error(err))
		} else {
			logg.Info("UpdateBook success", zap.Int("id", updatedBook.ID), zap.String("title", updatedBook.Title))
		}
	}

	if book != nil && reader != nil {
		reviews, _, err := reviewService.GetReviewsByBook(ctx, conn, book.ID, 10, 0)
		if err != nil {
			logg.Error("GetReviewsByBook error", zap.Int("book_id", book.ID), zap.Error(err))
		} else if len(reviews) > 0 {
			err = reviewService.DeleteReview(ctx, conn, reviews[0].ID, reader.Id)
			if err != nil {
				logg.Debug("DeleteReview error", zap.Int("review_id", reviews[0].ID), zap.Error(err))
			} else {
				logg.Info("DeleteReview success", zap.Int("review_id", reviews[0].ID))
			}
		}
	}

	settings, totalSettings, err := settingService.ListSettings(ctx, conn, 10, 0)
	if err != nil {
		logg.Error("ListSettings error", zap.Error(err))
	} else {
		logg.Info("ListSettings success", zap.Int("returned", len(settings)), zap.Int("total", totalSettings))
		for _, s := range settings {
			logg.Debug("ListSettings result", zap.String("key", s.Key), zap.String("value", s.Value))
		}
	}

}
