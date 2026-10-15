package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func IsUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
