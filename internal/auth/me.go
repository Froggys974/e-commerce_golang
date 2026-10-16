package auth

import (
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

func (handler *Handler) Me(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, ok := UserFromContext(r.Context())
	if !ok {
		httpjson.Error(w, http.StatusUnauthorized, "non authentifié")
		return
	}

	account, err := handler.users.GetUser(r.Context(), authenticatedUser.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, account)
}
