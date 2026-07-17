package handlers

import (
	"encoding/json"
	"errors"
	"library/internal/handlers/dto"
	"library/pkg/pagination"
	"net/http"
	"strconv"
	"strings"

	"library/internal/service"
	pkgerrors "library/pkg/errors"

	"github.com/gorilla/mux"
)

type BookHandler struct {
	service *service.BookService
}

func NewBookHandler(service *service.BookService) *BookHandler {
	return &BookHandler{service: service}
}

// 1. CreateBook — POST /api/v1/books

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	var req dto.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid_request", "Неверный формат запроса")
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	book := req.ToDomain()
	created, err := h.service.CreateBook(r.Context(), conn, &book, req.AuthorIDs, req.GenreIDs)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.BookResponseFromDomain(*created, req.AuthorIDs, req.GenreIDs)
	sendSuccess(w, http.StatusCreated, resp)
}

// 2. GetBook — GET /api/v1/books/{id}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		sendError(w, http.StatusBadRequest, "invalid_id", "Неверный ID книги")
		return
	}

	book, err := h.service.GetBook(r.Context(), conn, id)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.BookResponseFromDomain(*book, nil, nil)
	sendSuccess(w, http.StatusOK, resp)
}

// 3. GetBookByISBN — GET /api/v1/books/isbn/{isbn}

func (h *BookHandler) GetBookByISBN(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	isbn := vars["isbn"]
	if strings.TrimSpace(isbn) == "" {
		sendError(w, http.StatusBadRequest, "invalid_isbn", "ISBN не может быть пустым")
		return
	}

	book, err := h.service.GetBookByISBN(r.Context(), conn, isbn)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.BookResponseFromDomain(*book, nil, nil)
	sendSuccess(w, http.StatusOK, resp)
}

// 4. ListBooks — GET /api/v1/books

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pagination.LimitOffset(limit, offset)

	books, total, err := h.service.ListBooks(r.Context(), conn, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.NewBookListResponse(books, total, limit, offset, nil, nil)
	sendSuccess(w, http.StatusOK, resp)
}

// 5. SearchBooks — GET /api/v1/books/search

func (h *BookHandler) SearchBooks(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	column := r.URL.Query().Get("column")
	search := r.URL.Query().Get("search")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pagination.LimitOffset(limit, offset)

	if strings.TrimSpace(column) == "" {
		column = "title"
	}
	if strings.TrimSpace(search) == "" {
		sendError(w, http.StatusBadRequest, "empty_search", "Поисковый запрос не может быть пустым")
		return
	}

	books, total, err := h.service.SearchBooks(r.Context(), conn, column, search, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.NewBookListResponse(books, total, limit, offset, nil, nil)
	sendSuccess(w, http.StatusOK, resp)
}

// 6. GetPopularBooks — GET /api/v1/books/popular

func (h *BookHandler) GetPopularBooks(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	books, err := h.service.GetPopularBooks(r.Context(), conn, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.NewBookListResponse(books, len(books), limit, offset, nil, nil)
	sendSuccess(w, http.StatusOK, resp)
}

// 7. UpdateBook — PUT /api/v1/books/{id}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		sendError(w, http.StatusBadRequest, "invalid_id", "Неверный ID книги")
		return
	}

	var req dto.UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid_request", "Неверный формат запроса")
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.ISBN != nil {
		updates["isbn"] = *req.ISBN
	}
	if req.Year != nil {
		updates["year"] = *req.Year
	}
	if req.PublisherID != nil {
		updates["publisher_id"] = *req.PublisherID
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.CoverImageURL != nil {
		updates["cover_image_url"] = *req.CoverImageURL
	}

	book, err := h.service.UpdateBook(r.Context(), conn, id, updates, req.AuthorIDs, req.GenreIDs)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.BookResponseFromDomain(*book, req.AuthorIDs, req.GenreIDs)
	sendSuccess(w, http.StatusOK, resp)
}

// 8. DeleteBook — DELETE /api/v1/books/{id}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		sendError(w, http.StatusBadRequest, "invalid_id", "Неверный ID книги")
		return
	}

	if err := h.service.DeleteBook(r.Context(), conn, id, nil, nil); err != nil {
		sendServiceError(w, err)
		return
	}

	sendSuccess(w, http.StatusNoContent, nil)
}

// 9. AddCopy — POST /api/v1/books/{id}/copies

func (h *BookHandler) AddCopy(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["id"])
	if err != nil || bookID <= 0 {
		sendError(w, http.StatusBadRequest, "invalid_id", "Неверный ID книги")
		return
	}

	var req dto.CreateCopyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid_request", "Неверный формат запроса")
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	if err := h.service.AddCopyToBook(r.Context(), conn, bookID, req.Condition); err != nil {
		sendServiceError(w, err)
		return
	}

	sendSuccess(w, http.StatusCreated, map[string]interface{}{
		"book_id":   bookID,
		"condition": req.Condition,
		"message":   "Копия успешно добавлена",
	})
}

// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ

func sendSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.NewSuccessResponse(data))
}

func sendError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.NewErrorResponse(code, message))
}

func sendValidationError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	var details []dto.ValidationErrorDetail

	if ve, ok := err.(pkgerrors.ValidationErrors); ok {
		for _, e := range ve {
			details = append(details, dto.ValidationErrorDetail{
				Field:   e.Field,
				Message: e.Message,
			})
		}
	}

	json.NewEncoder(w).Encode(dto.NewValidationErrorResponse("validation_error", "Ошибка валидации", details))
}

func sendServiceError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"

	var notFound *pkgerrors.NotFoundError
	if errors.As(err, &notFound) {
		status = http.StatusNotFound
		code = "not_found"
	}

	var conflict *pkgerrors.ConflictError
	if errors.As(err, &conflict) {
		status = http.StatusConflict
		code = "conflict"
	}

	var validation *pkgerrors.ValidationError
	if errors.As(err, &validation) {
		status = http.StatusBadRequest
		code = "validation_error"
	}

	var business *pkgerrors.BusinessError
	if errors.As(err, &business) {
		status = http.StatusBadRequest
		code = business.Code
	}

	sendError(w, status, code, err.Error())
}
