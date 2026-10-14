package auth

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

type meResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (handler *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, found := UserFromContext(r.Context())
	if !found {
		httpjson.Error(w, http.StatusUnauthorized, "non authentifié")
		return
	}

	response := meResponse{ID: user.ID}
	err := handler.database.QueryRow(
		`SELECT email, role, created_at FROM users WHERE id = $1`,
		user.ID,
	).Scan(&response.Email, &response.Role, &response.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		httpjson.Error(w, http.StatusUnauthorized, "utilisateur introuvable")
		return
	}
	if err != nil {
		log.Printf("me: select: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	httpjson.Write(w, http.StatusOK, response)
}
