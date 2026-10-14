package auth

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

const invalidCredentialsMessage = "email ou mot de passe incorrect"

func (handler *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON non valide")
		return
	}

	email := normalizeEmail(request.Email)

	var (
		userID       int64
		passwordHash string
		role         string
		confirmed    bool
	)
	err := handler.database.QueryRow(
		`SELECT id, password_hash, role, confirmed FROM users WHERE email = $1`,
		email,
	).Scan(&userID, &passwordHash, &role, &confirmed)
	if errors.Is(err, sql.ErrNoRows) {
		httpjson.Error(w, http.StatusUnauthorized, invalidCredentialsMessage)
		return
	}
	if err != nil {
		log.Printf("login: select: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.Password))

	if err != nil {
		httpjson.Error(w, http.StatusUnauthorized, invalidCredentialsMessage)
		return
	}

	if !confirmed {
		httpjson.Error(w, http.StatusForbidden, "compte non confirmé")
		return
	}

	tokens, err := handler.issueTokens(userID, role)
	if err != nil {
		log.Printf("login: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}
	httpjson.Write(w, http.StatusOK, tokens)
}
