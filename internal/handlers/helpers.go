package handlers

import (
	"library/internal/middleware"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func getConnOrError(w http.ResponseWriter, r *http.Request) (*pgx.Conn, bool) {
	conn, err := middleware.GetConnFromContext(r)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "db_error", err.Error())
		return nil, false
	}
	return conn, true
}
