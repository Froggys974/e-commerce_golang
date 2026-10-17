package auth

import (
	"errors"
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type resetRequest struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"new_password"`
}

func (handler *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var request emailRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	code, err := handler.users.RequestPasswordReset(r.Context(), request.Email)
	switch {
	case errors.Is(err, user.ErrEmailNotFound):
	case err != nil:
		writeError(w, err)
		return
	default:
		handler.sendCode(request.Email, "Réinitialisation du mot de passe", code)
	}
	httpjson.Write(w, http.StatusOK, map[string]string{"message": codeSentMessage})
}

func (handler *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var request resetRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	if err := handler.users.ResetPassword(r.Context(), request.Email, request.Code, request.Password); err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]string{"message": "mot de passe modifié"})
}
