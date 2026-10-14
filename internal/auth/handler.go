package auth

import (
	"database/sql"
	"strings"
)

type Handler struct {
	database  *sql.DB
	jwtSecret []byte
}

func NewHandler(database *sql.DB, jwtSecret []byte) *Handler {
	return &Handler{database: database, jwtSecret: jwtSecret}
}

const (
	minPasswordLength = 8
	maxPasswordLength = 72 // limité avec bcrypt
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidPassword(password string) bool {
	return len(password) >= minPasswordLength && len(password) <= maxPasswordLength
}
