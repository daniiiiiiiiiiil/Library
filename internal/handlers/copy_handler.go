package handlers

import (
	"encoding/json"
	"library/internal/handlers/dto"
	"library/internal/middleware"
	"library/internal/service"
	"library/pkg/pagination"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CopyHandler struct {
	service *service.CopyService
}

func NewHTTPHandlersCopy(service *service.CopyService) *CopyHandler {
	return &CopyHandler{service: service}
}

func (h *CopyHandler) GetCopy(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "Неверный ID копии")
		return
	}
	copy, err := h.service.GetCopy(r.Context(), conn, id)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.CopyResponseFromDomain(*copy)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *CopyHandler) GetCopiesByBook(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)

	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["bookId"])
	if err != nil || bookID <= 0 {
		sendError(w, http.StatusBadRequest, "invalid_book_id", "Неверный ID книги")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pagination.LimitOffset(limit, offset)

	copies, err := h.service.GetCopiesByBook(r.Context(), conn, bookID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	var respCopies []dto.CopyResponse
	for _, copy := range copies {
		respCopies = append(respCopies, dto.CopyResponseFromDomain(copy))
	}

	resp := dto.CopyListResponse{
		Copies:     respCopies,
		Pagination: pagination.NewPagination(len(respCopies), limit, offset),
	}

	sendSuccess(w, http.StatusOK, resp)
}

func (h *CopyHandler) GetCopiesAvaliables(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["bookId"])
	if err != nil || bookID <= 0 {
		sendError(w, http.StatusBadRequest, "invalid_book_id", "Неверный ID книги")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	copies, err := h.service.GetAvailableCopies(r.Context(), conn, bookID)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	var respCopiesAvaliables []dto.CopyResponse
	for _, copy := range copies {
		respCopiesAvaliables = append(respCopiesAvaliables, dto.CopyResponseFromDomain(copy))
	}

	total := len(respCopiesAvaliables)

	resp := dto.CopyListResponse{
		Copies:     respCopiesAvaliables,
		Pagination: pagination.NewPagination(total, limit, offset),
	}
	sendSuccess(w, http.StatusOK, resp)
}

func (h *CopyHandler) PUTCopies(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)

	vars := mux.Vars(r)
	copyID, err := strconv.Atoi(vars["CopyID"])
	if err != nil || copyID <= 0 {
		sendError(w, http.StatusBadRequest, "Invalid_ID", "Неверный ID копии")
		return
	}
	var req dto.UpdateCopyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса")
		return
	}
	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}
	updates := make(map[string]interface{})
	if req.Condition != nil {
		updates["condition"] = *req.Condition
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	copy, err := h.service.UpdateCopy(r.Context(), conn, copyID, updates)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.CopyResponseFromDomain(*copy)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *CopyHandler) DeleteCopies(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	vars := mux.Vars(r)
	copyID, err := strconv.Atoi(vars["CopyID"])
	if err != nil || copyID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID копии не может быть меньше или равен 0")
		return
	}
	deleteCopy, err := h.service.DeleteCopy(r.Context(), conn, copyID)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	sendSuccess(w, http.StatusOK, deleteCopy)
}

func (h *CopyHandler) PathCopyStatus(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	vars := mux.Vars(r)
	copyID, err := strconv.Atoi(vars["CopyID"])
	if err != nil || copyID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID копии не может быть пустым")
		return
	}

	var req dto.UpdateCopyStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Не верный формат запроса")
		return
	}
	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}
	copy, err := h.service.UpdateCopyStatus(r.Context(), conn, copyID, req.Status, req.ReaderID)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.CopyResponseFromDomain(*copy)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *CopyHandler) GETCountAvailableCopy(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	vars := mux.Vars(r)
	copyID, err := strconv.Atoi(vars["CopyID"])
	if err != nil || copyID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}
	count, err := h.service.CountAvailableCopies(r.Context(), conn, copyID)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	sendSuccess(w, http.StatusOK, count)
}
