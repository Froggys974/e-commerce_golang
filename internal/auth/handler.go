package auth

import "database/sql"

type Handler struct {
	database  *sql.DB
	jwtSecret []byte
}

func NewHandler(database *sql.DB, jwtSecret []byte) *Handler {
	return &Handler{database: database, jwtSecret: jwtSecret}
}
