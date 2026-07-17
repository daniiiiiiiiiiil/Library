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

type ReservationHandler struct {
	service       *service.ReservationService
	copyService   *service.CopyService
	bookService   *service.BookService
	readerService *service.ReaderService
}

func NewReservationHandler(
	service *service.ReservationService,
	copyService *service.CopyService,
	bookService *service.BookService,
	readerService *service.ReaderService,
) *ReservationHandler {
	return &ReservationHandler{
		service:       service,
		copyService:   copyService,
		bookService:   bookService,
		readerService: readerService,
	}
}

func (h *ReservationHandler) CreateReservationBook(w http.ResponseWriter, r *http.Request) {
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
	copyID, err := strconv.Atoi(vars["copy_id"])
	if err != nil || copyID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID копии не может быть меньше или равен 0")
		return
	}

	var req dto.CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса")
		return
	}
	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	result, err := h.service.ReserveBook(r.Context(), conn, copyID, readerID)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.ReservationResponseFromDomain(result.Reservation, result.BookTitle, result.ReaderName)
	sendSuccess(w, http.StatusCreated, resp)
}

func (h *ReservationHandler) DeleteReservation(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	reservationID, err := strconv.Atoi(vars["reservation_id"])
	if err != nil || reservationID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0"+err.Error())
		return
	}
	if err := h.service.CancelReservation(r.Context(), conn, reservationID); err != nil {
		sendServiceError(w, err)
		return
	}
	sendSuccess(w, http.StatusNoContent, nil)
}

func (h *ReservationHandler) GetReservationByID(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	reservationID, err := strconv.Atoi(vars["reservation_id"])
	if err != nil || reservationID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}

	result, err := h.service.GetReservation(r.Context(), conn, reservationID)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.ReservationResponseFromDomain(result.Reservation, result.BookTitle, result.ReaderName)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReservationHandler) GetActiveReservedBookReader(w http.ResponseWriter, r *http.Request) {
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

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	activeReservation, total, err := h.service.GetActiveReaderReservationsWithDetails(r.Context(), conn, readerID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.NewReservationListResponseWithDetails(activeReservation, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReservationHandler) GetActiveReservationsByCopy(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	copyID, err := strconv.Atoi(vars["copyId"])
	if err != nil || copyID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID копии не может быть меньше или равен 0")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	reservations, total, err := h.service.GetActiveCopyReservations(r.Context(), conn, copyID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	bookTitles := make(map[int]string)
	readerNames := make(map[int]string)

	for _, res := range reservations {
		copy, err := h.copyService.GetCopy(r.Context(), conn, res.CopyID)
		if err != nil {
			continue
		}
		book, err := h.bookService.GetBook(r.Context(), conn, copy.BookID)
		if err != nil {
			continue
		}
		bookTitles[res.CopyID] = book.Title

		reader, err := h.readerService.GetReader(r.Context(), conn, res.ReaderID)
		if err != nil {
			continue
		}
		readerNames[res.ReaderID] = reader.Name
	}

	resp := dto.NewReservationListResponse(reservations, bookTitles, readerNames, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *ReservationHandler) ProcessExpiredReservations(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	var req dto.ProcessExpiredRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса")
		return
	}
	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	processedCount, err := h.service.ProcessExpiredReservations(r.Context(), conn, req.Limit)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.ProcessExpiredResponse{
		ProcessedCount: processedCount,
	}
	sendSuccess(w, http.StatusOK, resp)
}
