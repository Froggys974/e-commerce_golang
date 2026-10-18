package apiclient

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func newRefreshServer(t *testing.T, refreshCalls *int, mutex *sync.Mutex) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer new-access" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "token invalide ou expiré"})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 1, "email": "alice@test.local"})
	})
	mux.HandleFunc("POST /auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		*refreshCalls++
		mutex.Unlock()

		var request struct {
			RefreshToken string `json:"refresh_token"`
		}
		json.NewDecoder(r.Body).Decode(&request)
		if request.RefreshToken != "old-refresh" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "token invalide ou expiré"})
			return
		}
		json.NewEncoder(w).Encode(tokensResponse{AccessToken: "new-access", RefreshToken: "new-refresh"})
	})
	return httptest.NewServer(mux)
}

func TestRefreshOnUnauthorized(t *testing.T) {
	var refreshCalls int
	var mutex sync.Mutex
	server := newRefreshServer(t, &refreshCalls, &mutex)
	defer server.Close()

	client := New(server.URL)
	client.setTokens(tokensResponse{AccessToken: "old-access", RefreshToken: "old-refresh"})

	account, err := client.Me()
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if account.Email != "alice@test.local" {
		t.Errorf("Email = %q; attendu alice@test.local", account.Email)
	}
	if client.currentAccessToken() != "new-access" || client.refreshToken != "new-refresh" {
		t.Errorf("tokens = %q / %q; attendu les nouveaux", client.accessToken, client.refreshToken)
	}
	if refreshCalls != 1 {
		t.Errorf("refresh appelé %d fois; attendu 1", refreshCalls)
	}
}

func TestConcurrentUnauthorizedRefreshOnce(t *testing.T) {
	var refreshCalls int
	var mutex sync.Mutex
	server := newRefreshServer(t, &refreshCalls, &mutex)
	defer server.Close()

	client := New(server.URL)
	client.setTokens(tokensResponse{AccessToken: "old-access", RefreshToken: "old-refresh"})

	var waitGroup sync.WaitGroup
	errs := make(chan error, 5)
	for range 5 {
		waitGroup.Go(func() {
			_, err := client.Me()
			errs <- err
		})
	}
	waitGroup.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("erreur inattendue : %v", err)
		}
	}
	if refreshCalls != 1 {
		t.Errorf("refresh appelé %d fois; attendu 1", refreshCalls)
	}
}

func TestRefreshFailureKeepsUnauthorized(t *testing.T) {
	var refreshCalls int
	var mutex sync.Mutex
	server := newRefreshServer(t, &refreshCalls, &mutex)
	defer server.Close()

	client := New(server.URL)
	client.setTokens(tokensResponse{AccessToken: "old-access", RefreshToken: "revoked"})

	_, err := client.Me()

	var apiError *APIError
	if !errors.As(err, &apiError) || apiError.Status != http.StatusUnauthorized {
		t.Fatalf("erreur = %v; attendu un 401", err)
	}
}
