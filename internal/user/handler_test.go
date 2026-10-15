package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(User{Email: "florent@test.com", Role: RoleUser})
	handler := NewHandler(newTestService(repo))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/users", handler.List)
	mux.HandleFunc("POST /admin/users", handler.Create)
	mux.HandleFunc("PUT /admin/users/{id}", handler.Update)
	mux.HandleFunc("DELETE /admin/users/{id}", handler.Delete)
	mux.HandleFunc("POST /admin/users/{id}/confirm", handler.Confirm)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{"liste", "GET", "/admin/users", "", http.StatusOK},
		{"création", "POST", "/admin/users", `{"email":"bob@test.com","password":"motdepasse","role":"user"}`, http.StatusCreated},
		{"création invalide", "POST", "/admin/users", `{"email":"pas-un-email","password":"motdepasse","role":"user"}`, http.StatusBadRequest},
		{"modification", "PUT", "/admin/users/1", `{"email":"florent@test.com","role":"admin"}`, http.StatusOK},
		{"confirmation", "POST", "/admin/users/1/confirm", "", http.StatusOK},
		{"id invalide", "DELETE", "/admin/users/abc", "", http.StatusBadRequest},
		{"utilisateur inconnu", "DELETE", "/admin/users/999", "", http.StatusNotFound},
		{"suppression", "DELETE", "/admin/users/1", "", http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()

			mux.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d; attendu %d (%s)", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
		})
	}
}
