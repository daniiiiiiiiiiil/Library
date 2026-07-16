package dto

import (
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
		Error:     nil,
		Timestamp: time.Now(),
	}
}

func NewErrorResponse(code, message string) APIResponse {
	return APIResponse{
		Success: false,
		Data:    nil,
		Error: &ErrorResponse{
			Code:      code,
			Message:   message,
			Details:   nil,
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}
}

func NewValidationErrorResponse(code, message string, details []ValidationErrorDetail) APIResponse {
	return APIResponse{
		Success: false,
		Data:    nil,
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

type Pagination struct {
	Total       int    `json:"total"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
	CurrentPage int    `json:"current_page"`
	TotalPages  int    `json:"total_pages"`
	Next        string `json:"next,omitempty"`
	Previous    string `json:"previous,omitempty"`
}

func NewPagination(total, limit, offset int) Pagination {
	if limit <= 0 {
		limit = 10
	}
	currentPage := (offset / limit) + 1
	totalPages := (total + limit - 1) / limit

	return Pagination{
		Total:       total,
		Limit:       limit,
		Offset:      offset,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}
}

type HealthCheckResponse struct {
	Status   string    `json:"status"`
	Database string    `json:"database"`
	Uptime   string    `json:"uptime"`
	Version  string    `json:"version"`
	Time     time.Time `json:"time"`
}

func NewHealthCheckResponse(status, database, uptime, version string) HealthCheckResponse {
	return HealthCheckResponse{
		Status:   status,
		Database: database,
		Uptime:   uptime,
		Version:  version,
		Time:     time.Now(),
	}
}
