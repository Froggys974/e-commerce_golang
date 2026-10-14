package auth

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (handler *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	var userID int64
	err := handler.database.QueryRow(
		`DELETE FROM refresh_tokens
                   WHERE token_hash = $1 AND expires_at > now()
                   RETURNING user_id`,
		hashToken(request.RefreshToken),
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		httpjson.Error(w, http.StatusUnauthorized, "refresh token invalide ou expiré")
		return
	}
	if err != nil {
		log.Printf("refresh: delete: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	var role string
	err = handler.database.QueryRow(
		`SELECT role FROM users WHERE id = $1`,
		userID,
	).Scan(&role)
	if err != nil {
		log.Printf("refresh: select role: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	tokens, err := handler.issueTokens(userID, role)
	if err != nil {
		log.Printf("refresh: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}
	httpjson.Write(w, http.StatusOK, tokens)
}
