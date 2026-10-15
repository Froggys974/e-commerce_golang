package user

import (
	"errors"
	"fmt"
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func (r Role) Valid() bool {
	return r == RoleUser || r == RoleAdmin
}

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	Confirmed    bool      `json:"confirmed"`
	CreatedAt    time.Time `json:"created_at"`
}

var (
	ErrInvalidCode        = errors.New("code invalide ou expiré")
	ErrEmailNotFound      = errors.New("email introuvable")
	ErrInvalidCredentials = errors.New("email ou mot de passe incorrect")
	ErrNotConfirmed       = errors.New("compte non confirmé")
)

type NotFoundError struct {
	ID int64
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("utilisateur %d introuvable", e.ID)
}

type EmailTakenError struct {
	Email string
}

func (e EmailTakenError) Error() string {
	return fmt.Sprintf("email %s déjà utilisé", e.Email)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s : %s", e.Field, e.Message)
}
