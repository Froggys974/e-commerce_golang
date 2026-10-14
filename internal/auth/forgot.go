package auth

import (
	"crypto/rand"
	"log"
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

const forgotMessage = "Si un compte existe pour cet email, un code de réinitialisation a été envoyé"

type forgotRequest struct {
	Email string `json:"email"`
}

func (handler *Handler) Forgot(w http.ResponseWriter, r *http.Request) {
	var request forgotRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	email := normalizeEmail(request.Email)
	resetCode := rand.Text()

	result, err := handler.database.Exec(
		`UPDATE users
                   SET reset_code = $1, reset_expires_at = now() + interval '15 minutes'
                   WHERE email = $2`,
		hashToken(resetCode), email,
	)
	if err != nil {
		log.Printf("forgot: update: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("forgot: rows affected: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	if rowsAffected == 1 {
		// TODO : envoyer par email (mailpit / mailtrap ?)
		log.Printf("[MAIL] code de réinitialisation pour %s : %s", email, resetCode)
	}
	httpjson.Write(w, http.StatusOK, map[string]string{"message": forgotMessage})
}
