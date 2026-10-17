package auth

import (
	"errors"
	"log"
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// meme reponse si email deja pris anti enumeration
const codeSentMessage = "Si l'email est valide, un code a été envoyé"

func (handler *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	code, err := handler.users.Register(r.Context(), request.Email, request.Password)

	var emailTakenErr user.EmailTakenError
	switch {
	case errors.As(err, &emailTakenErr):
		log.Printf("register: %v", err)
	case err != nil:
		writeError(w, err)
		return
	default:
		handler.sendCode(request.Email, "Confirmation de votre compte", code)
	}
	httpjson.Write(w, http.StatusCreated, map[string]string{"message": codeSentMessage})
}
