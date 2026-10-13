package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/mail"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (handler *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	email := strings.ToLower(strings.TrimSpace(request.Email))
	parsedAddress, err := mail.ParseAddress(email)
	if err != nil || parsedAddress.Address != email {
		httpjson.Error(w, http.StatusBadRequest, "Si l'email est valide, un mail a été envoyé")
		return
	}

	if len(request.Password) < 8 || len(request.Password) > 72 {
		httpjson.Error(w, http.StatusBadRequest, "Le mot de passe doit faire entre 8 et 72 octs")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("register: bcrypt: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	confirmCode := generateConfirmCode()

	_, err = handler.database.Exec(
		`INSERT INTO users (email, password_hash, confirm_code) VALUES ($1, $2, $3)`,
		email, string(passwordHash), confirmCode,
	)

	if isUniqueViolation(err) {
		log.Printf("register: email déjà utilisé: %s", email)
		httpjson.Write(w, http.StatusCreated, map[string]string{
			"message": "Si l'email est valide, un code de confirmation a été envoyé",
		})
		return
	}

	if err != nil {
		log.Printf("register: insert: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	// TODO: envoyer par email via Mailpit / Mailtrap ?
	log.Printf("[MAIL] code de confirmation pour %s : %s", email, confirmCode)

	httpjson.Write(w, http.StatusCreated, map[string]string{
		"message": "Si l'email est valide, un code de confirmation a été envoyé",
	})
}

func generateConfirmCode() string {
	number, _ := rand.Int(rand.Reader, big.NewInt(1_000_000))
	return fmt.Sprintf("%06d", number.Int64())
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
