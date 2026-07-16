package handlers

import (
	"library/internal/service"
)

type HTTPHandlersTransaction struct {
	service *service.TransactionService
}

func NewHTTPHandlersTransaction(service *service.TransactionService) *HTTPHandlersTransaction {
	return &HTTPHandlersTransaction{
		service: service,
	}
}
