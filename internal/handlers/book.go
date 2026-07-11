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

type HTTPHandlerBook struct {
	book *domain.Book
}

func NewHTTPHandlerBook(book *domain.Book) *HTTPHandlerBook {
	return &HTTPHandlerBook{book: book}
}

func (handler *HTTPHandlerBook) HandlerCreateBook(w http.ResponseWriter, r *http.Request) {
	var bookDto dto.BookDTO
	if err := json.NewDecoder(r.Body).Decode(&bookDto); err != nil {
		errDTO := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
	}
	if err := bookDto.ValidateBook(); err != nil {
		errDTO := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
	}
	BookNew := models.NewBook(bookDto.Id, bookDto.Title, bookDto.Author, bookDto.Year, bookDto.IsAvailable, bookDto.ReaderId)
	if err := handler.book.AddBook(BookNew); err != nil {
		errDTO := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		if errors.Is(err, errorsMy.ErrBookAlreadyExists) {
			http.Error(w, errDTO.ToString(), http.StatusConflict)
		} else {
			http.Error(w, errDTO.ToString(), http.StatusInternalServerError)
		}
		return
	}
	b, err := json.MarshalIndent(BookNew, "", "    ")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Не удалось добавить книгу:", err)
		return
	}
}
func (handler *HTTPHandlerBook) HandlerGetBook(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}
	book, err := handler.book.GetBook(id)
	if err != nil {
		errDto := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		if errors.Is(err, errorsMy.ErrBookNotFound) {
			http.Error(w, errDto.ToString(), http.StatusBadRequest)
		} else {
			http.Error(w, errDto.ToString(), http.StatusInternalServerError)
		}
		return
	}
	b, err := json.MarshalIndent(book, "", "    ")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Не удалось найти задачу:", err)
	}
}

func (handler *HTTPHandlerBook) HandlerGetAllBooks(w http.ResponseWriter, r *http.Request) {
	books := handler.book.GetBooks()
	b, err := json.MarshalIndent(books, "", "    ")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Не удалось записать в http:", err)
		return
	}
}

func (handler *HTTPHandlerBook) HandlerUpdateBook(w http.ResponseWriter, r *http.Request) {
	var bookDto dto.BookDTO
	if err := json.NewDecoder(r.Body).Decode(&bookDto); err != nil {
		errDTO := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
		return
	}
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
	}
	BookNew := models.NewBook(id, bookDto.Title, bookDto.Author, bookDto.Year, bookDto.IsAvailable, bookDto.ReaderId)

	if err := bookDto.ValidateBook(); err != nil {
		errDTO := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
	}

	if err := handler.book.UpdateBook(id, BookNew); err != nil {
		errDto := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		if errors.Is(err, errorsMy.ErrBookNotFound) {
			http.Error(w, errDto.ToString(), http.StatusNotFound)
		} else {
			http.Error(w, errDto.ToString(), http.StatusInternalServerError)
		}
		return
	}
	b, err := json.MarshalIndent(BookNew, " ", "    ")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Не удалось обновить данные:", err)
		return
	}
}

func (handler *HTTPHandlerBook) HandlerDeleteBook(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Не удалось преобразовать id", http.StatusBadRequest)
	}
	if err := handler.book.DeleteBook(id); err != nil {
		errDto := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		if errors.Is(err, errorsMy.ErrBookAlreadyExists) {
			http.Error(w, errDto.ToString(), http.StatusNotFound)
		} else {
			http.Error(w, errDto.ToString(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
