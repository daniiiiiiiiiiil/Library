package handlers

import (
	"encoding/json"
	"library/internal/handlers/dto"
	"library/internal/middleware"
	"library/internal/service"
	"library/pkg/pagination"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type HTTPHandlersAuthor struct {
	service *service.AuthorService
}

func NewHTTPHandlersAuthor(service *service.AuthorService) *HTTPHandlersAuthor {
	return &HTTPHandlersAuthor{service: service}
}

func (h *HTTPHandlersAuthor) PostCreateAuthor(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	var req dto.CreateAuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса")
		return
	}
	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	author := req.ToDomain()
	created, err := h.service.CreateAuthor(r.Context(), conn, &author)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.AuthorResponseFromDomain(*created)
	sendSuccess(w, http.StatusCreated, resp)
}

func (h *HTTPHandlersAuthor) GetAuthors(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	authors, total, err := h.service.ListAuthors(r.Context(), conn, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.NewAuthorListResponse(authors, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *HTTPHandlersAuthor) GetAuthor(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	vars := mux.Vars(r)
	authorID, err := strconv.Atoi(vars["id"])
	if err != nil || authorID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}
	author, err := h.service.GetAuthor(r.Context(), conn, authorID)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.AuthorResponseFromDomain(*author)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *HTTPHandlersAuthor) UpdateAuthor(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	vars := mux.Vars(r)
	authorID, err := strconv.Atoi(vars["id"])
	if err != nil || authorID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}

	var req dto.UpdateAuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Не верный формат запроса")
		return
	}
	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}
	updates := make(map[string]interface{})
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Biography != nil {
		updates["biography"] = *req.Biography
	}
	if req.BirthDate != nil {
		updates["birth_date"] = *req.BirthDate
	}

	author, err := h.service.UpdateAuthor(r.Context(), conn, authorID, updates)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.AuthorResponseFromDomain(*author)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *HTTPHandlersAuthor) DeleteAuthor(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	vars := mux.Vars(r)
	authorID, err := strconv.Atoi(vars["id"])
	if err != nil || authorID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}
	if err := h.service.DeleteAuthor(r.Context(), conn, authorID); err != nil {
		sendServiceError(w, err)
		return
	}
	sendSuccess(w, http.StatusNoContent, nil)
}

func (h *HTTPHandlersAuthor) GetAuthorSearch(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)

	column := r.URL.Query().Get("column")
	search := r.URL.Query().Get("search")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pagination.LimitOffset(limit, offset)

	if strings.TrimSpace(column) == "" {
		column = "first_name"
	}
	if strings.TrimSpace(search) == "" {
		sendError(w, http.StatusBadRequest, "empty_search", "Поисковый запрос не может быть пустым")
		return
	}

	authors, total, err := h.service.SearchAuthors(r.Context(), conn, column, search, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.NewAuthorListResponse(authors, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *HTTPHandlersAuthor) AuthorsByBook(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["id"])
	if err != nil || bookID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}

	authors, total, err := h.service.GetAuthorsByBook(r.Context(), conn, bookID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.NewAuthorListResponse(authors, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}
