package handlers

import (
	"context"
	"encoding/json"
	"library/internal/domain"
	"library/internal/handlers/dto"
	"library/internal/middleware"
	"library/internal/service"
	"library/pkg/pagination"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type GenreHandlers struct {
	service *service.GenreService
}

func NewGenreHandlers(service *service.GenreService) *GenreHandlers {
	return &GenreHandlers{service: service}
}

func (h *GenreHandlers) CreateGenre(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)

	var req dto.CreateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	genre := req.ToDomain()
	create, err := h.service.CreateGenre(r.Context(), conn, genre)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "CreateGenre", err.Error())
		return
	}
	resp := dto.FromDomainGenreResponse(*create)
	sendSuccess(w, http.StatusCreated, resp)
}

func (h *GenreHandlers) GetGenres(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	genres, total, err := h.service.ListGenres(r.Context(), conn, limit, offset)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "GetGenres", err.Error())
		return
	}
	resp := dto.NewGenreListResponse(genres, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *GenreHandlers) GetGenre(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	vars := mux.Vars(r)
	genreID, err := strconv.Atoi(vars["genreID"])
	if err != nil || genreID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0"+err.Error())
		return
	}
	genre, err := h.service.GetGenre(r.Context(), conn, genreID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "GetGenre", err.Error())
		return
	}
	resp := dto.FromDomainGenreResponse(*genre)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *GenreHandlers) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	vars := mux.Vars(r)
	genreID, err := strconv.Atoi(vars["genreID"])
	if err != nil || genreID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0"+err.Error())
		return
	}
	var req dto.UpdateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат ввода "+err.Error())
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
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}

	genre, err := h.service.UpdateGenre(r.Context(), conn, genreID, updates)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "UpdateGenre", err.Error())
		return
	}

	resp := dto.FromDomainGenreResponse(*genre)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *GenreHandlers) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)
	vars := mux.Vars(r)
	genreID, err := strconv.Atoi(vars["genreID"])
	if err != nil || genreID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0"+err.Error())
		return
	}
	genre, err := h.service.DeleteGenre(r.Context(), conn, genreID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "DeleteGenre", err.Error())
		return
	}
	resp := dto.FromDomainGenreResponse(*genre)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *GenreHandlers) GetGenreHierarchy(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)

	genres, err := h.service.GetRootGenres(r.Context(), conn)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	var hierarchy []dto.GenreHierarchyResponse
	for _, genre := range genres {
		hierarchy = append(hierarchy, buildGenreHierarchy(genre, r.Context(), conn, h.service))
	}

	sendSuccess(w, http.StatusOK, hierarchy)
}

func buildGenreHierarchy(genre domain.Genre, ctx context.Context, conn *pgx.Conn, service *service.GenreService) dto.GenreHierarchyResponse {
	subGenres, _, err := service.GetSubGenres(ctx, conn, genre.ID)
	if err != nil {
		subGenres = []domain.Genre{}
	}

	response := dto.GenreHierarchyResponse{
		ID:   genre.ID,
		Name: genre.Name,
	}

	for _, sub := range subGenres {
		response.Children = append(response.Children, buildGenreHierarchy(sub, ctx, conn, service))
	}

	return response
}

func (h *GenreHandlers) GetSubgenres(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)

	vars := mux.Vars(r)
	ParentID, err := strconv.Atoi(vars["parent_id"])
	if err != nil || ParentID <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0"+err.Error())
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	subGenre, total, err := h.service.GetSubGenres(r.Context(), conn, ParentID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "GetSubGenres", err.Error())
		return
	}
	resp := dto.NewGenreListResponse(subGenre, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *GenreHandlers) GenreSearch(w http.ResponseWriter, r *http.Request) {
	conn := middleware.GetConnFromContext(r)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	search := r.URL.Query().Get("search")
	column := r.URL.Query().Get("column")

	pagination.LimitOffset(limit, offset)

	searchGet, total, err := h.service.SearchGenres(r.Context(), conn, column, search, limit, offset)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "GenreSearch", err.Error())
		return
	}
	resp := dto.NewGenreListResponse(searchGet, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}
