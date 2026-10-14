package auth

import (
	"log"
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

func (handler *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	_, err := handler.database.Exec(
		`DELETE FROM refresh_tokens WHERE token_hash = $1`,
		hashToken(request.RefreshToken),
	)
	if err != nil {
		log.Printf("logout: delete: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
