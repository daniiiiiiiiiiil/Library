package middleware

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type contextKey string

const DBConnKey contextKey = "db_conn"

func DBConnection(conn *pgx.Conn) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), DBConnKey, conn)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetConnFromContext(r *http.Request) (*pgx.Conn, error) {
	conn, ok := r.Context().Value(DBConnKey).(*pgx.Conn)
	if !ok || conn == nil {
		return nil, ErrDBConnectionNotFound
	}
	return conn, nil
}

var ErrDBConnectionNotFound = &DBConnectionError{}

type DBConnectionError struct{}

func (e *DBConnectionError) Error() string {
	return "database connection not found in context"
}
