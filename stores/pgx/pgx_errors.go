package pgx

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func isUniqueViolation(err error) bool {
	return pgErrorCode(err) == "23505"
}
