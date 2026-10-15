package auth

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type userService interface {
	Register(ctx context.Context, email, password string) (string, error)
	ResendConfirmCode(ctx context.Context, email string) (string, error)
	ConfirmByCode(ctx context.Context, email, code string) error
	RequestPasswordReset(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, email, code, password string) error
	GetUser(ctx context.Context, id int64) (user.User, error)
}

type Handler struct {
	users userService
	auth  *Service
}

func NewHandler(users userService, auth *Service) *Handler {
	return &Handler{users: users, auth: auth}
}

func writeError(w http.ResponseWriter, err error) {
	var validationErr user.ValidationError
	var notFoundErr user.NotFoundError

	switch {
	case errors.As(err, &validationErr):
		httpjson.Error(w, http.StatusBadRequest, validationErr.Error())
	case errors.Is(err, user.ErrInvalidCode):
		httpjson.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, user.ErrInvalidCredentials), errors.Is(err, ErrInvalidToken):
		httpjson.Error(w, http.StatusUnauthorized, err.Error())
	case errors.As(err, &notFoundErr):
		httpjson.Error(w, http.StatusUnauthorized, notFoundErr.Error())
	case errors.Is(err, user.ErrNotConfirmed):
		httpjson.Error(w, http.StatusForbidden, err.Error())
	default:
		log.Printf("auth: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
	}
}
