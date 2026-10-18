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

func (role Role) Valid() bool {
	return role == RoleUser || role == RoleAdmin
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

func (err NotFoundError) Error() string {
	return fmt.Sprintf("utilisateur %d introuvable", err.ID)
}

type EmailTakenError struct {
	Email string
}

func (err EmailTakenError) Error() string {
	return fmt.Sprintf("email %s déjà utilisé", err.Email)
}

type ValidationError struct {
	Field   string
	Message string
}

func (err ValidationError) Error() string {
	return fmt.Sprintf("%s : %s", err.Field, err.Message)
}
