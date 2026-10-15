package auth

import (
	"errors"
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type confirmRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type emailRequest struct {
	Email string `json:"email"`
}

func (handler *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	var request confirmRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	if err := handler.users.ConfirmByCode(r.Context(), request.Email, request.Code); err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]string{"message": "compte confirmé"})
}

func (handler *Handler) ResendConfirmCode(w http.ResponseWriter, r *http.Request) {
	var request emailRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	code, err := handler.users.ResendConfirmCode(r.Context(), request.Email)
	switch {
	case errors.Is(err, user.ErrEmailNotFound):
	case err != nil:
		writeError(w, err)
		return
	default:
		sendCode(request.Email, code)
	}
	httpjson.Write(w, http.StatusOK, map[string]string{"message": codeSentMessage})
}
