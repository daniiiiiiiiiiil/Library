package domain

import (
	"fmt"
	"library/pkg/errors"
	"strings"
	"time"
)

type Book struct {
	ID            int        `json:"id"`
	Title         string     `json:"title"`
	ISBN          string     `json:"isbn"`
	Year          int        `json:"year"`
	PublisherID   int        `json:"publisher_id"`
	Description   string     `json:"description"`
	CoverImageURL string     `json:"cover_image_url"`
	AvgRating     float64    `json:"avg_rating"`
	ReviewsCount  int        `json:"reviews_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

func NewBook(
	bookID int,
	title string,
	isbn string,
	year int,
	publisherID int,
	description string,
	coverImageURL string,
	avgRating float64,
	reviewsCount int,
	createdAt time.Time,
	updatedAt *time.Time,
) Book {
	return Book{
		ID:            bookID,
		Title:         title,
		ISBN:          isbn,
		Year:          year,
		PublisherID:   publisherID,
		Description:   description,
		CoverImageURL: coverImageURL,
		AvgRating:     avgRating,
		ReviewsCount:  reviewsCount,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

func (b Book) Validate() error {
	var errs errors.ValidationErrors

	if b.Title == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "title",
			Message: "название книги не может быть пустым",
		})
	} else if len(b.Title) > 200 {
		errs = append(errs, errors.ValidationError{
			Field:   "title",
			Message: fmt.Sprintf("название книги не может превышать 200 символов (сейчас %d)", len(b.Title)),
		})
	}

	if b.ISBN == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "isbn",
			Message: "ISBN не может быть пустым",
		})
	} else if !IsValidISBN(b.ISBN) {
		errs = append(errs, errors.ValidationError{
			Field:   "isbn",
			Message: "некорректный формат ISBN (должен быть 10 или 13 цифр)",
		})
	}

	if b.Year < 1400 {
		errs = append(errs, errors.ValidationError{
			Field:   "year",
			Message: fmt.Sprintf("год не может быть меньше 1400 (текущий: %d)", b.Year),
		})
	} else if b.Year > 2030 {
		errs = append(errs, errors.ValidationError{
			Field:   "year",
			Message: fmt.Sprintf("год не может быть больше 2030 (текущий: %d)", b.Year),
		})
	}

	if b.PublisherID <= 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "publisher_id",
			Message: "ID издателя должен быть положительным числом",
		})
	}

	if b.Description == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "description",
			Message: "описание книги не может быть пустым",
		})
	} else if len(b.Description) < 10 {
		errs = append(errs, errors.ValidationError{
			Field:   "description",
			Message: "описание должно содержать минимум 10 символов",
		})
	}

	if b.CoverImageURL == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "cover_image_url",
			Message: "URL обложки не может быть пустым",
		})
	} else if !IsValidURL(b.CoverImageURL) {
		errs = append(errs, errors.ValidationError{
			Field:   "cover_image_url",
			Message: "некорректный URL обложки",
		})
	}

	if b.AvgRating < 0 || b.AvgRating > 10 {
		errs = append(errs, errors.ValidationError{
			Field:   "avg_rating",
			Message: fmt.Sprintf("рейтинг должен быть от 0 до 10 (текущий: %.2f)", b.AvgRating),
		})
	}

	if b.ReviewsCount < 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "reviews_count",
			Message: "количество отзывов не может быть отрицательным",
		})
	}

	if errs.HasErrors() {
		return errs
	}

	return nil
}

func IsValidISBN(isbn string) bool {
	cleaned := strings.ReplaceAll(isbn, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	return len(cleaned) == 10 || len(cleaned) == 13
}

func IsValidURL(url string) bool {
	if url == "" {
		return false
	}
	return strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "/")
}

func (b *Book) Update(
	title string,
	isbn string,
	year int,
	publisherID int,
	description string,
	coverImageURL string,
) error {
	if title != "" {
		b.Title = title
	}
	if isbn != "" {
		b.ISBN = isbn
	}
	if year > 0 {
		b.Year = year
	}
	if publisherID > 0 {
		b.PublisherID = publisherID
	}
	if description != "" {
		b.Description = description
	}
	if coverImageURL != "" {
		b.CoverImageURL = coverImageURL
	}

	now := time.Now()
	b.UpdatedAt = &now

	return b.Validate()
}
