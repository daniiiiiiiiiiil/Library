package handlers

import (
	"encoding/json"
	"library/internal/handlers/dto"
	"library/internal/service"
	"library/pkg/pagination"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type ReaderHandlers struct {
	service *service.ReaderService
}

func NewReaderHandlers(service *service.ReaderService) *ReaderHandlers {
	return &ReaderHandlers{service: service}
}

func (h *ReaderHandlers) CreateReader(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	var req dto.CreateReaderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса")
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	reader := req.ToDomain()
	created, err := h.service.CreateReader(r.Context(), conn, &reader, req.Password)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.ReaderFromDomain(*created)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReaderHandlers) GetReader(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pagination.LimitOffset(limit, offset)

	reader, total, err := h.service.ListReader(r.Context(), conn, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.NewReaderListResponse(reader, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReaderHandlers) GetReaderID(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	ReaderID, err := strconv.Atoi(vars["id"])
	if err != nil || ReaderID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidId", "ID не может быть меньше или равен 0")
		return
	}
	reader, err := h.service.GetReader(r.Context(), conn, ReaderID)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.ReaderFromDomain(*reader)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReaderHandlers) GetReaderEmail(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	ReaderEmail := vars["email"]
	if strings.TrimSpace(ReaderEmail) == "" {
		sendError(w, http.StatusBadRequest, "InvalidEmail", "Email не может быть пустым")
	}

	reader, err := h.service.GetByEmail(r.Context(), conn, ReaderEmail)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.ReaderFromDomain(*reader)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReaderHandlers) UpdateReader(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	ReaderID, err := strconv.Atoi(vars["id"])
	if err != nil || ReaderID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть больше или равен 0")
		return
	}

	var req dto.UpdateReaderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса")
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.MaxBooks != nil {
		updates["maxBooks"] = *req.MaxBooks
	}

	reader, err := h.service.Update(r.Context(), conn, ReaderID, updates)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.ReaderFromDomain(*reader)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReaderHandlers) DeleteReader(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	ReaderID, err := strconv.Atoi(vars["id"])
	if err != nil || ReaderID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равно нулю")
		return
	}
	if err := h.service.Delete(r.Context(), conn, ReaderID); err != nil {
		sendServiceError(w, err)
		return
	}
	sendSuccess(w, http.StatusNoContent, nil)
}

func (h *ReaderHandlers) BlockReader(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	readerID, err := strconv.Atoi(vars["id"])
	if err != nil || readerID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}

	if err := h.service.BlockReader(r.Context(), conn, readerID); err != nil {
		sendServiceError(w, err)
		return
	}

	sendSuccess(w, http.StatusOK, map[string]interface{}{
		"reader_id": readerID,
		"status":    "blocked",
		"message":   "Читатель успешно заблокирован",
	})
}

func (h *ReaderHandlers) UnblockReader(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	readerID, err := strconv.Atoi(vars["id"])
	if err != nil || readerID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}

	if err := h.service.UnblockReader(r.Context(), conn, readerID); err != nil {
		sendServiceError(w, err)
		return
	}

	sendSuccess(w, http.StatusOK, map[string]interface{}{
		"reader_id": readerID,
		"status":    "active",
		"message":   "Читатель успешно разблокирован",
	})
}

func (h *ReaderHandlers) GetActiveReaders(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pagination.LimitOffset(limit, offset)

	readers, err := h.service.GetActiveReaders(r.Context(), conn, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.NewReaderListResponse(readers, len(readers), limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReaderHandlers) GetDebtors(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pagination.LimitOffset(limit, offset)

	readers, err := h.service.GetDebtors(r.Context(), conn, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.NewReaderListResponse(readers, len(readers), limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReaderHandlers) GetReaderHistory(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	readerID, err := strconv.Atoi(vars["id"])
	if err != nil || readerID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pagination.LimitOffset(limit, offset)

	transactions, total, err := h.service.GetReaderHistory(r.Context(), conn, readerID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	var transactionResponses []dto.TransactionResponse
	for _, tx := range transactions {
		transactionResponses = append(transactionResponses, dto.TransactionRequestFromDomain(tx))
	}

	resp := map[string]interface{}{
		"reader_id":    readerID,
		"transactions": transactionResponses,
		"pagination":   pagination.NewPagination(total, limit, offset),
	}

	sendSuccess(w, http.StatusOK, resp)
}
