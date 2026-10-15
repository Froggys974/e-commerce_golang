package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Froggys974/e-commerce_golang/internal/user"
)

func TestRequireAdmin(t *testing.T) {
	secret := []byte("secret")
	handler := NewHandler(nil, NewService(nil, nil, secret))

	protected := handler.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tokenFor := func(role user.Role) string {
		token, err := signAccessToken(secret, 1, role, time.Now())
		if err != nil {
			t.Fatalf("erreur inattendue : %v", err)
		}
		return "Bearer " + token
	}

	tests := []struct {
		name          string
		authorization string
		wantStatus    int
	}{
		{"sans token", "", http.StatusUnauthorized},
		{"token invalide", "Bearer abc", http.StatusUnauthorized},
		{"utilisateur", tokenFor(user.RoleUser), http.StatusForbidden},
		{"admin", tokenFor(user.RoleAdmin), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/admin/users", nil)
			request.Header.Set("Authorization", tt.authorization)
			recorder := httptest.NewRecorder()

			protected(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d; attendu %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
