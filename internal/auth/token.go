package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Froggys974/e-commerce_golang/internal/user"
)

var ErrInvalidToken = errors.New("token invalide ou expiré")

type infoToken struct {
	Role user.Role `json:"role"`
	jwt.RegisteredClaims
}

type AuthenticatedUser struct {
	ID   int64
	Role user.Role
}

func signAccessToken(secret []byte, userID int64, role user.Role, now time.Time) (string, error) {
	info := infoToken{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, info).SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("signature du token : %w", err)
	}
	return signed, nil
}

func parseAccessToken(secret []byte, token string, now time.Time) (AuthenticatedUser, error) {
	var info infoToken
	_, err := jwt.ParseWithClaims(token, &info,
		// any impose par golang-jwt
		func(*jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
	if err != nil {
		return AuthenticatedUser{}, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(info.Subject, 10, 64)
	if err != nil {
		return AuthenticatedUser{}, ErrInvalidToken
	}
	return AuthenticatedUser{ID: userID, Role: info.Role}, nil
}
