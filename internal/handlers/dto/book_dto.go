package dto

import (
	"library/internal/domain"
	"library/pkg/errors"
	"library/pkg/pagination"
	"strings"
	"time"
)

type CreateBookRequest struct {
	Title         string `json:"title"`
	ISBN          string `json:"isbn"`
	Year          int    `json:"year"`
	PublisherID   int    `json:"publisher_id"`
	Description   string `json:"description"`
	CoverImageURL string `json:"cover_image_url"`
	AuthorIDs     []int  `json:"author_ids"`
	GenreIDs      []int  `json:"genre_ids"`
}

func (r *CreateBookRequest) Validate() error {
	var errs errors.ValidationErrors

	if strings.TrimSpace(r.Title) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "title",
			Message: "название книги не может быть пустым",
		})
	}

	if strings.TrimSpace(r.ISBN) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "isbn",
			Message: "ISBN не может быть пустым",
		})
	}

	if r.Year < 1400 || r.Year > 2030 {
		errs = append(errs, errors.ValidationError{
			Field:   "year",
			Message: "год должен быть между 1400 и 2030",
		})
	}

	if r.PublisherID <= 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "publisher_id",
			Message: "ID издателя должен быть положительным",
		})
	}

	if strings.TrimSpace(r.Description) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "description",
			Message: "описание не может быть пустым",
		})
	}

	if len(r.AuthorIDs) == 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "author_ids",
			Message: "должен быть хотя бы один автор",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *CreateBookRequest) ToDomain() domain.Book {
	return domain.NewBook(
		0,
		r.Title,
		r.ISBN,
		r.Year,
		r.PublisherID,
		r.Description,
		r.CoverImageURL,
		0, // avg_rating — будет рассчитан позже
		0, // reviews_count — будет рассчитан позже
		time.Now(),
		nil, // updated_at — пока не обновляли
	)
}

type UpdateBookRequest struct {
	Title         *string `json:"title,omitempty"`
	ISBN          *string `json:"isbn,omitempty"`
	Year          *int    `json:"year,omitempty"`
	PublisherID   *int    `json:"publisher_id,omitempty"`
	Description   *string `json:"description,omitempty"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
	AuthorIDs     []int   `json:"author_ids,omitempty"`
	GenreIDs      []int   `json:"genre_ids,omitempty"`
}

func (r *UpdateBookRequest) Validate() error {
	var errs errors.ValidationErrors

	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "title",
			Message: "название не может быть пустым",
		})
	}

	if r.ISBN != nil && strings.TrimSpace(*r.ISBN) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "isbn",
			Message: "ISBN не может быть пустым",
		})
	}

	if r.Year != nil && (*r.Year < 1400 || *r.Year > 2030) {
		errs = append(errs, errors.ValidationError{
			Field:   "year",
			Message: "год должен быть между 1400 и 2030",
		})
	}

	if r.PublisherID != nil && *r.PublisherID <= 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "publisher_id",
			Message: "ID издателя должен быть положительным",
		})
	}

	if r.Description != nil && strings.TrimSpace(*r.Description) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "description",
			Message: "описание не может быть пустым",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *UpdateBookRequest) ToDomain(bookID int) (domain.Book, error) {
	if bookID <= 0 {
		return domain.Book{}, errors.NewBusinessError("invalid_id", "ID книги должен быть положительным")
	}

	return domain.NewBook(
		bookID,
		"",
		"",
		0,
		0,
		"",
		"",
		0,
		0,
		time.Now(),
		nil,
	), nil
}

type BookResponse struct {
	ID            int        `json:"ID"`
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
	AuthorIDs     []int      `json:"author_ids,omitempty"`
	GenreIDs      []int      `json:"genre_ids,omitempty"`
}

func BookResponseFromDomain(book domain.Book, authorIDs, genreIDs []int) BookResponse {
	return BookResponse{
		ID:            book.ID,
		Title:         book.Title,
		ISBN:          book.ISBN,
		Year:          book.Year,
		PublisherID:   book.PublisherID,
		Description:   book.Description,
		CoverImageURL: book.CoverImageURL,
		AvgRating:     book.AvgRating,
		ReviewsCount:  book.ReviewsCount,
		CreatedAt:     book.CreatedAt,
		UpdatedAt:     book.UpdatedAt,
		AuthorIDs:     authorIDs,
		GenreIDs:      genreIDs,
	}
}

type BookListResponse struct {
	Books      []BookResponse        `json:"books"`
	Pagination pagination.Pagination `json:"pagination"`
}

func NewBookListResponse(books []domain.Book, total, limit, offset int, authorIDs, genreIDs map[int][]int) BookListResponse {
	resp := BookListResponse{
		Books:      make([]BookResponse, 0, len(books)),
		Pagination: pagination.Calculate(total, limit, offset),
	}

	for _, book := range books {
		resp.Books = append(resp.Books, BookResponseFromDomain(
			book,
			authorIDs[book.ID],
			genreIDs[book.ID],
		))
	}

	return resp
}

type SearchBookRequest struct {
	Title   string `json:"title,omitempty"`
	Author  string `json:"author,omitempty"`
	Genre   string `json:"genre,omitempty"`
	YearMin int    `json:"year_min,omitempty"`
	YearMax int    `json:"year_max,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
}

func (r *SearchBookRequest) Validate() error {
	var errs errors.ValidationErrors

	if r.YearMin > 0 && r.YearMax > 0 && r.YearMin > r.YearMax {
		errs = append(errs, errors.ValidationError{
			Field:   "year_min",
			Message: "год начала не может быть больше года окончания",
		})
	}

	if r.Limit < 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "limit",
			Message: "limit не может быть отрицательным",
		})
	}

	if r.Offset < 0 {
		errs = append(errs, errors.ValidationError{
			Field:   "offset",
			Message: "offset не может быть отрицательным",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}
