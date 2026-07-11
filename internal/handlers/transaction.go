package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"library/errorsMy"
	"library/internal/domain"
	"library/internal/handlers/dto"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type HTTPHandlerTransaction struct {
	transaction *domain.Transaction
	book        *domain.Book
	reader      *domain.Reader
}

func NewHTTPHandlerTransaction(transaction *domain.Transaction, book *domain.Book, reader *domain.Reader) *HTTPHandlerTransaction {
	return &HTTPHandlerTransaction{transaction: transaction, book: book, reader: reader}
}

func (handler *HTTPHandlerTransaction) HandlerBorrowBook(w http.ResponseWriter, r *http.Request) {
	var transactionDto dto.BorrowReturnDTO
	if err := json.NewDecoder(r.Body).Decode(&transactionDto); err != nil {
		http.Error(w, "Не удалось прочитать с тела запроса", http.StatusBadRequest)
		return
	}
	book, err := handler.book.GetBook(transactionDto.BookID)
	if err != nil {
		if errors.Is(err, errorsMy.ErrBookNotFound) {
			http.Error(w, "Книга не найдена", http.StatusNotFound)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if !book.IsAvailable {
		http.Error(w, "Книга уже выдана", http.StatusConflict)
		return
	}

	_, err = handler.reader.GetReader(transactionDto.ReaderID)
	if err != nil {
		if errors.Is(err, errorsMy.ErrReaderNotFound) {
			http.Error(w, "Читатель не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	tx := models.NewTransaction(transactionDto.TransactionID, transactionDto.BookID,
		transactionDto.ReaderID, "borrow")
	if err := handler.transaction.AddTransactions(tx.ID, tx); err != nil {
		http.Error(w, "Не удалось добавить транзакцию", http.StatusBadRequest)
		return
	}
	if err := handler.book.UpdateBookStatus(transactionDto.BookID, false, transactionDto.ReaderID); err != nil {
		http.Error(w, "Не удалось обновить статус книги", http.StatusInternalServerError)
		return
	}

	if err := handler.reader.IncrementReaderBookCount(transactionDto.ReaderID); err != nil {
		http.Error(w, "Не удалось обновить счетчик читателя", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}

func (handler *HTTPHandlerTransaction) HandlerReturnBook(w http.ResponseWriter, r *http.Request) {
	var transactionDto dto.BorrowReturnDTO

	if err := json.NewDecoder(r.Body).Decode(&transactionDto); err != nil {
		http.Error(w, "Не удалось прочитать с тела запроса", http.StatusBadRequest)
		return
	}

	book, err := handler.book.GetBook(transactionDto.BookID)
	if err != nil {
		if errors.Is(err, errorsMy.ErrBookNotFound) {
			http.Error(w, "Книга не найдена", http.StatusBadRequest)
			return
		} else {
			http.Error(w, "Внутряняя ошибка сервера", http.StatusInternalServerError)
			return
		}
	}
	if book.IsAvailable {
		http.Error(w, "Книга не выдана", http.StatusConflict)
		return
	}

	_, err = handler.reader.GetReader(transactionDto.ReaderID)
	if err != nil {
		if errors.Is(err, errorsMy.ErrReaderNotFound) {
			http.Error(w, "Читатель не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	tx := models.NewTransaction(transactionDto.TransactionID, transactionDto.BookID,
		transactionDto.ReaderID, "return")

	if err := handler.transaction.AddTransactions(tx.ID, tx); err != nil {
		http.Error(w, "Не удалось добавить транзакцию", http.StatusBadRequest)
		return
	}

	if err := handler.book.UpdateBookStatus(transactionDto.BookID, true, 0); err != nil {
		http.Error(w, "Не удалось обновить статус книги", http.StatusInternalServerError)
		return
	}

	if err := handler.reader.DecrementReaderBookCount(transactionDto.ReaderID); err != nil {
		http.Error(w, "Не удалось обновить счетчик читателя", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tx)
}

func (handler *HTTPHandlerTransaction) HandlerGetAllTransactions(w http.ResponseWriter, r *http.Request) {
	transaction := handler.transaction.GetAllTransactions()
	b, err := json.MarshalIndent(transaction, " ", "    ")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Не удалось записать в http:", err)
		return
	}
}

func (handler *HTTPHandlerTransaction) HandlerGetBookTransactions(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Не удалось преобразовать id в int:", err)
		return
	}
	bookTransaction, err := handler.transaction.GetBookTransactions(id)
	if err != nil {
		errDto := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		if errors.Is(err, errorsMy.ErrReaderHasBook) {
			http.Error(w, errDto.ToString(), http.StatusBadRequest)
		} else {
			http.Error(w, errDto.ToString(), http.StatusInternalServerError)
		}
		return
	}
	b, err := json.MarshalIndent(bookTransaction, "", "    ")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Не удалось записать:", err)
	}
}
func (handler *HTTPHandlerTransaction) HandlerGetReaderTransactions(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Не удалось преобразовать id в int:", err)
		return
	}
	readerTransaction, err := handler.transaction.GetReaderTransactions(id)
	if err != nil {
		errDto := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		if errors.Is(err, errorsMy.ErrReaderNotFound) {
			http.Error(w, errDto.ToString(), http.StatusBadRequest)
		} else {
			http.Error(w, errDto.ToString(), http.StatusInternalServerError)
		}
		return
	}
	b, err := json.MarshalIndent(readerTransaction, " ", "    ")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Не удалось записать:", err)
	}
}
