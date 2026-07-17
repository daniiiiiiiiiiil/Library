package handlers

import (
	"encoding/json"
	"library/internal/handlers/dto"
	"library/internal/service"
	"library/pkg/pagination"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type TransactionHandler struct {
	service       *service.TransactionService
	readerService *service.ReaderService
	copyService   *service.CopyService
	bookService   *service.BookService
}

func NewTransactionHandler(service *service.TransactionService, readerService *service.ReaderService, copyService *service.CopyService, bookService *service.BookService) *TransactionHandler {
	return &TransactionHandler{
		service:       service,
		readerService: readerService,
		copyService:   copyService,
		bookService:   bookService,
	}
}

func (h *TransactionHandler) BorrowBook(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	var req dto.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	dueDate := req.DueDate
	if dueDate.IsZero() {
		dueDate = time.Now().Add(14 * 24 * time.Hour)
	}

	transaction, err := h.service.BorrowBook(r.Context(), conn, req.CopyID, req.ReaderID, dueDate)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.TransactionRequestFromDomain(*transaction)
	sendSuccess(w, http.StatusCreated, resp)
}

func (h *TransactionHandler) ReturnBook(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	var req dto.ReturnBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	returnDate := req.ReturnDate
	if returnDate.IsZero() {
		returnDate = time.Now()
	}

	fine := req.FineAmount

	transaction, err := h.service.ReturnBook(r.Context(), conn, req.TransactionID, returnDate, fine)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.TransactionRequestFromDomain(*transaction)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	transactionID, err := strconv.Atoi(vars["transaction_id"])
	if err != nil && transactionID != 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0"+err.Error())
		return
	}
	transaction, err := h.service.GetTransaction(r.Context(), conn, transactionID)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	resp := dto.TransactionRequestFromDomain(*transaction)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *TransactionHandler) GetTransactionByReader(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	vars := mux.Vars(r)
	readerID, err := strconv.Atoi(vars["reader_id"])
	if err != nil && readerID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0"+err.Error())
		return
	}
	transactionReader, total, err := h.service.GetReaderTransactions(r.Context(), conn, readerID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}
	bookTitles := make(map[int]string)
	for _, tx := range transactionReader {
		if _, ok := bookTitles[tx.CopyID]; ok {
			continue
		}
		copy, err := h.copyService.GetCopy(r.Context(), conn, tx.CopyID)
		if err != nil {
			continue
		}
		book, err := h.bookService.GetBook(r.Context(), conn, copy.BookID)
		if err != nil {
			continue
		}
		bookTitles[tx.CopyID] = book.Title
	}
	readerNames := make(map[int]string)
	for _, tx := range transactionReader {
		if _, ok := readerNames[tx.ReaderID]; ok {
			continue
		}
		reader, err := h.readerService.GetReader(r.Context(), conn, tx.ReaderID)
		if err != nil {
			continue
		}
		readerNames[tx.ReaderID] = reader.Name
	}
	resp := dto.NewTransactionListResponse(transactionReader, bookTitles, readerNames, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *TransactionHandler) GetActiveReaderTransactions(w http.ResponseWriter, r *http.Request) {
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

	transactions, err := h.service.GetActiveReaderTransactions(r.Context(), conn, readerID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	bookTitles := make(map[int]string)
	for _, tx := range transactions {
		if _, ok := bookTitles[tx.CopyID]; ok {
			continue
		}
		copy, err := h.copyService.GetCopy(r.Context(), conn, tx.CopyID)
		if err != nil {
			continue
		}
		book, err := h.bookService.GetBook(r.Context(), conn, copy.BookID)
		if err != nil {
			continue
		}
		bookTitles[tx.CopyID] = book.Title
	}

	readerNames := make(map[int]string)
	for _, tx := range transactions {
		if _, ok := readerNames[tx.ReaderID]; ok {
			continue
		}
		reader, err := h.readerService.GetReader(r.Context(), conn, tx.ReaderID)
		if err != nil {
			continue
		}
		readerNames[tx.ReaderID] = reader.Name
	}

	resp := dto.NewTransactionListResponse(transactions, bookTitles, readerNames, len(transactions), limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *TransactionHandler) GetOverdueTransactions(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	transactions, err := h.service.GetOverdueTransactions(r.Context(), conn, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	bookTitles := make(map[int]string)
	for _, tx := range transactions {
		if _, ok := bookTitles[tx.CopyID]; ok {
			continue
		}
		copy, err := h.copyService.GetCopy(r.Context(), conn, tx.CopyID)
		if err != nil {
			continue
		}
		book, err := h.bookService.GetBook(r.Context(), conn, copy.BookID)
		if err != nil {
			continue
		}
		bookTitles[tx.CopyID] = book.Title
	}

	readerNames := make(map[int]string)
	for _, tx := range transactions {
		if _, ok := readerNames[tx.ReaderID]; ok {
			continue
		}
		reader, err := h.readerService.GetReader(r.Context(), conn, tx.ReaderID)
		if err != nil {
			continue
		}
		readerNames[tx.ReaderID] = reader.Name
	}

	resp := dto.NewTransactionListResponse(transactions, bookTitles, readerNames, len(transactions), limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *TransactionHandler) ProcessOverdueTransactions(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	var req dto.ProcessOverdueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса")
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	if err := h.service.ProcessOverdueTransactions(r.Context(), conn, req.Limit, 0); err != nil {
		sendServiceError(w, err)
		return
	}

	sendSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Просроченные транзакции успешно обработаны",
	})
}

func (h *TransactionHandler) GetBookTransactions(w http.ResponseWriter, r *http.Request) {
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

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	transactions, total, err := h.service.GetBookTransactions(r.Context(), conn, bookID, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	bookTitles := make(map[int]string)
	for _, tx := range transactions {
		if _, ok := bookTitles[tx.CopyID]; ok {
			continue
		}
		copy, err := h.copyService.GetCopy(r.Context(), conn, tx.CopyID)
		if err != nil {
			continue
		}
		book, err := h.bookService.GetBook(r.Context(), conn, copy.BookID)
		if err != nil {
			continue
		}
		bookTitles[tx.CopyID] = book.Title
	}

	readerNames := make(map[int]string)
	for _, tx := range transactions {
		if _, ok := readerNames[tx.ReaderID]; ok {
			continue
		}
		reader, err := h.readerService.GetReader(r.Context(), conn, tx.ReaderID)
		if err != nil {
			continue
		}
		readerNames[tx.ReaderID] = reader.Name
	}

	resp := dto.NewTransactionListResponse(transactions, bookTitles, readerNames, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}
