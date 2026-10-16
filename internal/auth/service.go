package auth

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/Froggys974/e-commerce_golang/internal/ref"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type users interface {
	Authenticate(ctx context.Context, email, password string) (user.User, error)
	GetUser(ctx context.Context, id int64) (user.User, error)
}

type tokenRepository interface {
	CreateToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	RotateToken(ctx context.Context, oldHash, newHash string, expiresAt time.Time) (int64, error)
	RevokeToken(ctx context.Context, tokenHash string) error
}

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type Service struct {
	users  users
	tokens tokenRepository
	secret []byte
}

func NewService(userService users, tokenRepo tokenRepository, secret []byte) *Service {
	return &Service{users: userService, tokens: tokenRepo, secret: secret}
}

func (s *Service) Login(ctx context.Context, email, password string) (Tokens, error) {
	account, err := s.users.Authenticate(ctx, email, password)
	if err != nil {
		return Tokens{}, err
	}

	refreshToken := rand.Text()
	if err := s.tokens.CreateToken(ctx, account.ID, ref.Hash(refreshToken), time.Now().Add(refreshTokenTTL)); err != nil {
		return Tokens{}, err
	}
	return s.newTokens(account, refreshToken)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (Tokens, error) {
	newRefreshToken := rand.Text()
	userID, err := s.tokens.RotateToken(ctx, ref.Hash(refreshToken), ref.Hash(newRefreshToken), time.Now().Add(refreshTokenTTL))
	if err != nil {
		return Tokens{}, err
	}

	account, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return Tokens{}, err
	}
	return s.newTokens(account, newRefreshToken)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.tokens.RevokeToken(ctx, ref.Hash(refreshToken))
}

func (s *Service) newTokens(account user.User, refreshToken string) (Tokens, error) {
	accessToken, err := signAccessToken(s.secret, account.ID, account.Role, time.Now())
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(accessTokenTTL.Seconds()),
	}, nil
}

func (s *Service) parseAccessToken(token string) (AuthenticatedUser, error) {
	return parseAccessToken(s.secret, token, time.Now())
}
