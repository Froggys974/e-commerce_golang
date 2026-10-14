package auth

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

type resetRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (handler *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	var request resetRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	if !isValidPassword(request.NewPassword) {
		httpjson.Error(w, http.StatusBadRequest, "Le mot de passe doit faire entre 8 et 72 octets")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("reset: bcrypt: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	var userID int64
	err = handler.database.QueryRow(
		`UPDATE users
                   SET password_hash = $1, reset_code = NULL, reset_expires_at = NULL
                   WHERE email = $2 AND reset_code = $3 AND reset_expires_at > now()
                   RETURNING id`,
		string(passwordHash), normalizeEmail(request.Email), hashToken(strings.TrimSpace(request.Code)),
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		httpjson.Error(w, http.StatusBadRequest, "code invalide ou expiré")
		return
	}
	if err != nil {
		log.Printf("reset: update: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	_, err = handler.database.Exec(`DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	if err != nil {
		log.Printf("reset: revoke sessions: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]string{
		"message": "mot de passe modifié, veuillez vous reconnecter",
	})
}
