package dto

import (
	"library/internal/domain"
	"library/pkg/errors"
	"library/pkg/pagination"
	"strings"
	"time"
)

type CreateSettingRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func (r *CreateSettingRequest) Validate() error {
	var errs errors.ValidationErrors

	if strings.TrimSpace(r.Key) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "key",
			Message: "Ключ не может быть пустым",
		})
	}

	if strings.TrimSpace(r.Value) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "value",
			Message: "Значение не может быть пустым",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *CreateSettingRequest) ToDomain() domain.Setting {
	return domain.NewSetting(
		0,
		r.Key,
		r.Value,
		r.Description,
		nil,
	)
}

type UpdateSettingRequest struct {
	Value       *string `json:"value,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (r *UpdateSettingRequest) Validate() error {
	var errs errors.ValidationErrors

	if r.Value != nil && strings.TrimSpace(*r.Value) == "" {
		errs = append(errs, errors.ValidationError{
			Field:   "value",
			Message: "Значение не может быть пустым",
		})
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

type SettingResponse struct {
	ID          int        `json:"id"`
	Key         string     `json:"key"`
	Value       string     `json:"value"`
	Description string     `json:"description"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func SettingResponseFromDomain(s domain.Setting) SettingResponse {
	return SettingResponse{
		ID:          s.ID,
		Key:         s.Key,
		Value:       s.Value,
		Description: s.Description,
		UpdatedAt:   s.UpdatedAt,
	}
}

type SettingListResponse struct {
	Settings   []SettingResponse     `json:"settings"`
	Pagination pagination.Pagination `json:"pagination"`
}

func NewSettingListResponse(settings []domain.Setting, total, limit, offset int) SettingListResponse {
	resp := SettingListResponse{
		Settings:   make([]SettingResponse, 0, len(settings)),
		Pagination: pagination.NewPagination(total, limit, offset),
	}

	for _, s := range settings {
		resp.Settings = append(resp.Settings, SettingResponseFromDomain(s))
	}

	return resp
}
