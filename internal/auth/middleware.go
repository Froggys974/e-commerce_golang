package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type contextKey struct{}

func (handler *Handler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			httpjson.Error(w, http.StatusUnauthorized, "token manquant")
			return
		}

		authenticatedUser, err := handler.auth.parseAccessToken(token)
		if err != nil {
			httpjson.Error(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), contextKey{}, authenticatedUser)
		next(w, r.WithContext(ctx))
	}
}

func (handler *Handler) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return handler.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		authenticatedUser, _ := UserFromContext(r.Context())
		if authenticatedUser.Role != user.RoleAdmin {
			httpjson.Error(w, http.StatusForbidden, "accès réservé aux admins")
			return
		}
		next(w, r)
	})
}

func UserFromContext(ctx context.Context) (AuthenticatedUser, bool) {
	authenticatedUser, ok := ctx.Value(contextKey{}).(AuthenticatedUser)
	return authenticatedUser, ok
}
