package handlers

import (
	"encoding/json"
	"library/internal/handlers/dto"
	"library/internal/service"
	"library/pkg/pagination"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ReviewHandlers struct {
	service       *service.ReviewService
	bookService   *service.BookService
	readerService *service.ReaderService
}

func NewReviewHandlers(service *service.ReviewService, bookService *service.BookService, readerService *service.ReaderService) *ReviewHandlers {
	return &ReviewHandlers{
		service:       service,
		bookService:   bookService,
		readerService: readerService,
	}
}

func (h *ReviewHandlers) CreateReview(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	readerID, err := strconv.Atoi(vars["reader_id"])
	if err != nil || readerID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID читателя не может быть меньше или равен 0")
		return
	}

	var req dto.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidID", "Неверный формат запроса")
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	review := req.ToDomain(readerID)
	created, err := h.service.CreateReview(r.Context(), conn, review)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	book, err := h.bookService.GetBook(r.Context(), conn, created.BookID)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	reader, err := h.readerService.GetReader(r.Context(), conn, created.ReaderID)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.ReviewRequestFromDomain(*created, book.Title, reader.Name)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReviewHandlers) GetReview(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	reviewID, err := strconv.Atoi(vars["review_id"])
	if err != nil || reviewID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0"+err.Error())
		return
	}
	review, err := h.service.GetReview(r.Context(), conn, reviewID)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	book, err := h.bookService.GetBook(r.Context(), conn, review.BookID)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	reader, err := h.readerService.GetReader(r.Context(), conn, review.ReaderID)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.ReviewRequestFromDomain(*review, book.Title, reader.Name)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReviewHandlers) GetReviewBook(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["book_id"])
	if err != nil || bookID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID книги не может быть меньше или равен 0")
		return
	}

	reviews, total, err := h.service.GetReviewsByBook(r.Context(), conn, bookID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	avgRating, err := h.service.GetAverageRating(r.Context(), conn, bookID)
	if err != nil {
		avgRating = 0
	}

	book, err := h.bookService.GetBook(r.Context(), conn, bookID)
	bookTitle := ""
	if err == nil {
		bookTitle = book.Title
	}

	readerNames := make(map[int]string)
	for _, review := range reviews {
		reader, err := h.readerService.GetReader(r.Context(), conn, review.ReaderID)
		if err != nil {
			continue
		}
		readerNames[review.ReaderID] = reader.Name
	}

	bookTitles := make(map[int]string)
	bookTitles[bookID] = bookTitle

	resp := dto.NewReviewListResponse(reviews, bookTitles, readerNames, total, limit, offset, avgRating)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReviewHandlers) GetReviewReader(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	vars := mux.Vars(r)
	readerID, err := strconv.Atoi(vars["reader_id"])
	if err != nil || readerID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}

	reviews, total, err := h.service.GetReviewsByReader(r.Context(), conn, readerID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	bookTitles := make(map[int]string)
	for _, review := range reviews {
		if _, ok := bookTitles[review.BookID]; ok {
			continue
		}
		book, err := h.bookService.GetBook(r.Context(), conn, review.BookID)
		if err != nil {
			continue
		}
		bookTitles[review.BookID] = book.Title
	}

	readerNames := make(map[int]string)
	for _, review := range reviews {
		if _, ok := readerNames[review.ReaderID]; ok {
			continue
		}
		reader, err := h.readerService.GetReader(r.Context(), conn, review.ReaderID)
		if err != nil {
			continue
		}
		readerNames[review.ReaderID] = reader.Name
	}

	avgRating := 0.0
	if len(reviews) > 0 {
		var sum float64
		for _, review := range reviews {
			sum += review.Rating
		}
		avgRating = sum / float64(len(reviews))
	}

	resp := dto.NewReviewListResponse(reviews, bookTitles, readerNames, total, limit, offset, avgRating)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReviewHandlers) GetReviewsByReader(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	readerID, err := strconv.Atoi(vars["readerId"])
	if err != nil || readerID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID читателя не может быть меньше или равен 0")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	reviews, total, err := h.service.GetReviewsByReader(r.Context(), conn, readerID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	bookTitles := make(map[int]string)
	for _, review := range reviews {
		book, err := h.bookService.GetBook(r.Context(), conn, review.BookID)
		if err != nil {
			continue
		}
		bookTitles[review.BookID] = book.Title
	}

	readerNames := make(map[int]string)
	for _, review := range reviews {
		reader, err := h.readerService.GetReader(r.Context(), conn, review.ReaderID)
		if err != nil {
			continue
		}
		readerNames[review.ReaderID] = reader.Name
	}

	resp := dto.NewReviewListResponse(reviews, bookTitles, readerNames, total, limit, offset, 0)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReviewHandlers) UpdateReview(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	reviewID, err := strconv.Atoi(vars["id"])
	if err != nil || reviewID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID отзыва не может быть меньше или равен 0")
		return
	}

	var req dto.UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса")
		return
	}
	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	// Получаем readerID из контекста (JWT)
	// и в других файлах тоже это есть
	readerID := 0 // TODO: Получить из JWT

	rating := 0.0
	if req.Rating != nil {
		rating = *req.Rating
	}
	comment := ""
	if req.Comment != nil {
		comment = *req.Comment
	}

	updatedReview, err := h.service.UpdateReview(
		r.Context(),
		conn,
		reviewID,
		readerID,
		rating,
		comment,
	)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	book, _ := h.bookService.GetBook(r.Context(), conn, updatedReview.BookID)
	bookTitle := ""
	if book != nil {
		bookTitle = book.Title
	}

	reader, _ := h.readerService.GetReader(r.Context(), conn, updatedReview.ReaderID)
	readerName := ""
	if reader != nil {
		readerName = reader.Name
	}

	resp := dto.ReviewRequestFromDomain(*updatedReview, bookTitle, readerName)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReviewHandlers) DeleteReview(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	reviewID, err := strconv.Atoi(vars["id"])
	if err != nil || reviewID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID отзыва не может быть меньше или равен 0")
		return
	}

	// Получаем ID текущего пользователя из контекста (JWT)
	readerID := 0 // TODO: Получить из JWT

	if err := h.service.DeleteReview(r.Context(), conn, reviewID, readerID); err != nil {
		sendServiceError(w, err)
		return
	}

	sendSuccess(w, http.StatusNoContent, nil)
}

func (h *ReviewHandlers) GetBookRating(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["bookId"])
	if err != nil || bookID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID книги не может быть меньше или равен 0")
		return
	}

	avgRating, reviewsCount, err := h.service.GetBookRating(r.Context(), conn, bookID)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	book, _ := h.bookService.GetBook(r.Context(), conn, bookID)
	bookTitle := ""
	if book != nil {
		bookTitle = book.Title
	}

	resp := dto.BookRatingResponse{
		BookID:        bookID,
		BookTitle:     bookTitle,
		AverageRating: avgRating,
		ReviewsCount:  reviewsCount,
	}

	sendSuccess(w, http.StatusOK, resp)
}
