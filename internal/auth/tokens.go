package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (handler *Handler) newAccessToken(userID int64, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenDuration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(handler.jwtSecret)
}

func (handler *Handler) newRefreshToken(userID int64) (string, error) {
	token := rand.Text()
	_, err := handler.database.Exec(
		`INSERT INTO refresh_tokens (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		hashToken(token), userID, time.Now().Add(refreshTokenDuration),
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func hashToken(token string) string {
	hashBytes := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hashBytes[:])
}

type tokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (handler *Handler) issueTokens(userID int64, role string) (tokensResponse, error) {
	accessToken, err := handler.newAccessToken(userID, role)
	if err != nil {
		return tokensResponse{}, fmt.Errorf("access token: %w", err)
	}

	refreshToken, err := handler.newRefreshToken(userID)
	if err != nil {
		return tokensResponse{}, fmt.Errorf("refresh token: %w", err)
	}

	return tokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(accessTokenDuration.Seconds()),
	}, nil
}
