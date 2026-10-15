package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/Froggys974/e-commerce_golang/internal/user"
)

func TestAccessToken(t *testing.T) {
	secret := []byte("secret")
	now := time.Now()

	token, err := signAccessToken(secret, 42, user.RoleAdmin, now)
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}

	authenticatedUser, err := parseAccessToken(secret, token, now)
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if authenticatedUser.ID != 42 || authenticatedUser.Role != user.RoleAdmin {
		t.Errorf("user = %+v; attendu ID 42 et Role admin", authenticatedUser)
	}

	tests := []struct {
		name   string
		secret []byte
		token  string
		now    time.Time
	}{
		{"mauvais secret", []byte("autre"), token, now},
		{"token modifié", secret, token + "x", now},
		{"token expiré", secret, token, now.Add(accessTokenTTL + time.Minute)},
		{"pas un token", secret, "abc", now},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseAccessToken(tt.secret, tt.token, tt.now)
			if !errors.Is(err, ErrInvalidToken) {
				t.Errorf("erreur = %v; attendu ErrInvalidToken", err)
			}
		})
	}
}
