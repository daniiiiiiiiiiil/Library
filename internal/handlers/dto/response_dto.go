package dto

import (
	"library/pkg/pagination"
	"time"
)

type APIResponse struct {
	Success   bool           `json:"success"`
	Data      interface{}    `json:"data,omitempty"`
	Error     *ErrorResponse `json:"error,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}
}

func NewErrorResponse(code, message string, details []ValidationErrorDetail) APIResponse {
	return APIResponse{
		Success: false,
		Error: &ErrorResponse{
			Code:      code,
			Message:   message,
			Details:   details,
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}
}

type ErrorResponse struct {
	Code      string                  `json:"code"`
	Message   string                  `json:"message"`
	Details   []ValidationErrorDetail `json:"details,omitempty"`
	Timestamp time.Time               `json:"timestamp"`
}

type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type PaginationResponse struct {
	Total       int    `json:"total"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
	CurrentPage int    `json:"current_page"`
	TotalPages  int    `json:"total_pages"`
	Next        string `json:"next,omitempty"`
	Previous    string `json:"previous,omitempty"`
}

func PaginationResponseFromDomain(p pagination.Pagination, next, previous string) PaginationResponse {
	return PaginationResponse{
		Total:       p.Total,
		Limit:       p.Limit,
		Offset:      p.Offset,
		CurrentPage: p.CurrentPage,
		TotalPages:  p.TotalPages,
		Next:        next,
		Previous:    previous,
	}
}

type HealthCheckResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Uptime   string `json:"uptime"`
	Version  string `json:"version"`
}

func NewHealthCheckResponse(status, dbStatus, uptime, version string) HealthCheckResponse {
	return HealthCheckResponse{
		Status:   status,
		Database: dbStatus,
		Uptime:   uptime,
		Version:  version,
	}
}
