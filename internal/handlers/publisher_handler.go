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

type PublisherHandler struct {
	service *service.PublisherService
}

func NewPublisherHandler(service *service.PublisherService) *PublisherHandler {
	return &PublisherHandler{
		service: service,
	}
}

func (h *PublisherHandler) CreatePublisher(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	var req dto.CreatePublisherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}
	publisher := req.ToDomain()
	create, err := h.service.CreatePublisher(r.Context(), conn, &publisher)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.PublisherResponseFromDomain(*create)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *PublisherHandler) GetPublishers(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	publisher, total, err := h.service.ListPublishers(r.Context(), conn, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	req := dto.NewPublisherListResponse(publisher, total, limit, offset)
	sendSuccess(w, http.StatusOK, req)
}

func (h *PublisherHandler) GetPublisher(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	PublisherID, err := strconv.Atoi(vars["publisherId"])
	if err != nil || PublisherID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равно 0"+err.Error())
		return
	}
	publisher, err := h.service.GetPublisher(r.Context(), conn, PublisherID)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.PublisherResponseFromDomain(*publisher)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *PublisherHandler) UpdatePublisher(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	PublisherID, err := strconv.Atoi(mux.Vars(r)["publisherId"])
	if err != nil || PublisherID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат ввода")
		return
	}
	var req dto.UpdatePublisherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса"+err.Error())
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
	if req.Address != nil {
		updates["address"] = *req.Address
	}

	publisher, err := h.service.UpdatePublisher(r.Context(), conn, PublisherID, updates)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.PublisherResponseFromDomain(*publisher)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *PublisherHandler) DeletePublisher(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	PublisherID, err := strconv.Atoi(vars["publisherId"])
	if err != nil || PublisherID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равно 0"+err.Error())
		return
	}
	if err := h.service.DeletePublisher(r.Context(), conn, PublisherID); err != nil {
		sendServiceError(w, err)
		return
	}

	sendSuccess(w, http.StatusNoContent, nil)
}

func (h *PublisherHandler) SearchPublisher(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	search := r.URL.Query().Get("search")
	column := r.URL.Query().Get("column")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	searchPublisher, total, err := h.service.SearchPublishers(r.Context(), conn, column, search, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	req := dto.NewPublisherListResponse(searchPublisher, total, limit, offset)
	sendSuccess(w, http.StatusOK, req)
}

func (h *PublisherHandler) GetBooksPublisher(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	publisherID, err := strconv.Atoi(vars["publisherId"])
	if err != nil || publisherID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0"+err.Error())
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	books, total, err := h.service.GetPublisherBooks(r.Context(), conn, publisherID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.NewBookListResponse(books, total, limit, offset, nil, nil)
	sendSuccess(w, http.StatusOK, resp)
}
