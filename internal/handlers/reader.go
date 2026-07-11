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

type HTTPHandlerReader struct {
	reader *domain.Reader
}

func NewHTTPHandlerReader(reader *domain.Reader) *HTTPHandlerReader {
	return &HTTPHandlerReader{reader: reader}
}

func (handler *HTTPHandlerReader) HandlerCreateReader(w http.ResponseWriter, r *http.Request) {
	var readerDto dto.ReaderDTO
	if err := json.NewDecoder(r.Body).Decode(&readerDto); err != nil {
		errDto := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		http.Error(w, errDto.ToString(), http.StatusBadRequest)
		return
	}
	if err := readerDto.ValidateReader(); err != nil {
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
	ReaderNew := models.NewReader(readerDto.Id, readerDto.Name, readerDto.Phone, readerDto.Email, readerDto.BookCount)
	if err := handler.reader.AddReader(ReaderNew); err != nil {
		errDto := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		if errors.Is(err, errorsMy.ErrReaderNotFound) {
			http.Error(w, errDto.ToString(), http.StatusConflict)
		} else {
			http.Error(w, errDto.ToString(), http.StatusInternalServerError)
		}
		return
	}
	b, err := json.MarshalIndent(ReaderNew, "", "    ")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Не удалось записать в http", err)
		return
	}

}

func (handler *HTTPHandlerReader) HandlerGetReader(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Не удалось преобразовать id в чисто:", err)
		return
	}
	reader, err := handler.reader.GetReader(id)
	if err != nil {
		errDTO := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		if errors.Is(err, errorsMy.ErrReaderNotFound) {
			http.Error(w, errDTO.ToString(), http.StatusBadRequest)
		} else {
			http.Error(w, errDTO.ToString(), http.StatusInternalServerError)
		}
		return
	}
	b, err := json.MarshalIndent(reader, "", "    ")
	if err != nil {
		panic(err)
	}
	if _, err := w.Write(b); err != nil {
		fmt.Println("Не удалось записать в http", err)
	}
}
func (handler *HTTPHandlerReader) HandlerGetAllReaders(w http.ResponseWriter, r *http.Request) {
	reader := handler.reader.GetAllReaders()
	b, err := json.MarshalIndent(reader, "", "    ")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Не удалось отправить запрос http:", err)
		return
	}
}

func (hadler *HTTPHandlerReader) HandlerUpdateReader(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Не удалось преобразовать id", http.StatusBadRequest)
	}
	var readerDTO dto.ReaderDTO
	if err := json.NewDecoder(r.Body).Decode(&readerDTO); err != nil {
		errDto := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		if errors.Is(err, errorsMy.ErrReaderNotFound) {
			http.Error(w, errDto.ToString(), http.StatusBadRequest)
		} else {
			http.Error(w, errDto.ToString(), http.StatusInternalServerError)
		}
	}
	ReaderNew := models.NewReader(id, readerDTO.Name, readerDTO.Phone, readerDTO.Email, readerDTO.BookCount)
	if err := readerDTO.ValidateReader(); err != nil {
		errDto := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		http.Error(w, errDto.ToString(), http.StatusBadRequest)
	}
	if err := hadler.reader.UpdateReader(id, ReaderNew); err != nil {
		errDto := dto.ErrorDTO{
			Message: err.Error(),
			Time:    time.Now(),
		}
		if errors.Is(err, errorsMy.ErrReaderNotFound) {
			http.Error(w, errDto.ToString(), http.StatusBadRequest)
		} else {
			http.Error(w, errDto.ToString(), http.StatusBadRequest)
		}
		return
	}
	b, err := json.MarshalIndent(ReaderNew, "", "    ")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Не удалось записать ")
		return
	}
}

func (handler *HTTPHandlerReader) HandlerDeleteReader(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Не удалось преобразовать id в int:", err)
		return
	}
	if err := handler.reader.DeleteReader(id); err != nil {
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
	w.WriteHeader(http.StatusNoContent)
}
