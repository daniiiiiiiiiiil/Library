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

type SettingsHandler struct {
	service *service.SettingService
}

func NewSettingsHandler(service *service.SettingService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pagination.LimitOffset(limit, offset)

	settings, total, err := h.service.ListSettings(r.Context(), conn, limit, offset)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.NewSettingListResponse(settings, total, limit, offset)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *SettingsHandler) GetSettingByID(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}

	setting, err := h.service.GetSetting(r.Context(), conn, id)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.SettingResponseFromDomain(*setting)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *SettingsHandler) GetSettingByKey(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		sendError(w, http.StatusBadRequest, "InvalidKey", "Ключ не может быть пустым")
		return
	}

	setting, err := h.service.GetSettingByKey(r.Context(), conn, key)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.SettingResponseFromDomain(*setting)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *SettingsHandler) CreateSetting(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}
	var req dto.CreateSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	setting := req.ToDomain()
	created, err := h.service.CreateSetting(r.Context(), conn, &setting)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.SettingResponseFromDomain(*created)
	sendSuccess(w, http.StatusCreated, resp)
}

func (h *SettingsHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}

	var req dto.UpdateSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	updates := make(map[string]interface{})
	if req.Value != nil {
		updates["value"] = *req.Value
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	setting, err := h.service.UpdateSetting(r.Context(), conn, id, updates)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.SettingResponseFromDomain(*setting)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *SettingsHandler) UpdateSettingByKey(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		sendError(w, http.StatusBadRequest, "InvalidKey", "Ключ не может быть пустым")
		return
	}

	var req dto.UpdateSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "InvalidRequest", "Неверный формат запроса: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		sendValidationError(w, err)
		return
	}

	if req.Value == nil {
		sendError(w, http.StatusBadRequest, "InvalidValue", "Значение обязательно для обновления")
		return
	}

	if err := h.service.UpdateSettingByKey(r.Context(), conn, key, *req.Value); err != nil {
		sendServiceError(w, err)
		return
	}

	setting, err := h.service.GetSettingByKey(r.Context(), conn, key)
	if err != nil {
		sendServiceError(w, err)
		return
	}

	resp := dto.SettingResponseFromDomain(*setting)
	sendSuccess(w, http.StatusOK, resp)
}

func (h *SettingsHandler) DeleteSetting(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		sendError(w, http.StatusBadRequest, "InvalidID", "ID не может быть меньше или равен 0")
		return
	}

	if err := h.service.DeleteSetting(r.Context(), conn, id); err != nil {
		sendServiceError(w, err)
		return
	}

	sendSuccess(w, http.StatusNoContent, nil)
}

func (h *SettingsHandler) DeleteSettingByKey(w http.ResponseWriter, r *http.Request) {
	conn, ok := getConnOrError(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		sendError(w, http.StatusBadRequest, "InvalidKey", "Ключ не может быть пустым")
		return
	}

	if err := h.service.DeleteSettingByKey(r.Context(), conn, key); err != nil {
		sendServiceError(w, err)
		return
	}

	sendSuccess(w, http.StatusNoContent, nil)
}
