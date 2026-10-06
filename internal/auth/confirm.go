package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

type confirmRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (handler *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	var request confirmRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	email := strings.ToLower(strings.TrimSpace(request.Email))
	code := strings.TrimSpace(request.Code)

	result, err := handler.database.Exec(
		`UPDATE users
        SET confirmed = true, confirm_code = NULL
        WHERE email = $1 AND confirm_code = $2 AND confirmed = false`,
		email, code,
	)
	if err != nil {
		log.Printf("confirm: update: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("confirm: rows affected: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}
	if rowsAffected == 0 {
		httpjson.Error(w, http.StatusBadRequest, "code invalide ou compte déjà confirmé")
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]string{
		"message": "compte confirmé",
	})
}
