package auth

import (
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (handler *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	tokens, err := handler.auth.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, tokens)
}

func (handler *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	tokens, err := handler.auth.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, tokens)
}

func (handler *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	if err := handler.auth.Logout(r.Context(), request.RefreshToken); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
