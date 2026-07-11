package main

import (
	"fmt"
	handlers2 "library/internal/handlers"
)

func main() {
	books := storage.NewBooks()
	readers := storage.NewReaders()
	transactions := storage.NewTransactions()
	httpHandlersBooks := handlers2.NewHTTPHandlerBook(books)
	httpHandlersReaders := handlers2.NewHTTPHandlerReader(readers)
	httpHandlersTransactions := handlers2.NewHTTPHandlerTransaction(transactions, books, readers)
	httpServer := server.NewHTTPServer(httpHandlersBooks, httpHandlersReaders, httpHandlersTransactions)
	if err := httpServer.Start(); err != nil {
		fmt.Println("Не удалось запустить http сервер:", err)
	}
}
