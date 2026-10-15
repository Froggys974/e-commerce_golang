package auth

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

type contextKey string

const userContextKey contextKey = "user"

type AuthenticatedUser struct {
	ID   int64
	Role string
}

func (handler *Handler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		tokenString, hasBearerPrefix := strings.CutPrefix(authorizationHeader, "Bearer ")
		if !hasBearerPrefix || tokenString == "" {
			httpjson.Error(w, http.StatusUnauthorized, "token manquant")
			return
		}

		claims := &Claims{}
		_, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(token *jwt.Token) (any, error) {
				return handler.jwtSecret, nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		)
		if err != nil {
			httpjson.Error(w, http.StatusUnauthorized, "token invalide ou expiré")
			return
		}

		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			httpjson.Error(w, http.StatusUnauthorized, "token invalide ou expiré")
			return
		}

		user := AuthenticatedUser{ID: userID, Role: claims.Role}
		contextWithUser := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(contextWithUser))
	}
}

func UserFromContext(requestContext context.Context) (AuthenticatedUser, bool) {
	user, found := requestContext.Value(userContextKey).(AuthenticatedUser)
	return user, found
}

func (handler *Handler) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return handler.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := UserFromContext(r.Context())
		if user.Role != "admin" {
			httpjson.Error(w, http.StatusForbidden, "User non admin")
			return
		}
		next(w, r)
	})
}
