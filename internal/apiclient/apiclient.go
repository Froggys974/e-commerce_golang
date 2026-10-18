package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client

	mutex        sync.Mutex
	accessToken  string
	refreshToken string
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type APIError struct {
	Status  int
	Message string
}

func (apiError *APIError) Error() string {
	return apiError.Message
}

type tokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (client *Client) setTokens(tokens tokensResponse) {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	client.accessToken = tokens.AccessToken
	client.refreshToken = tokens.RefreshToken
}

func (client *Client) currentAccessToken() string {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return client.accessToken
}

// --> Les méthodes client passent ici <--
func (client *Client) do(method, path string, body any, destination any) error {
	var encoded []byte
	if body != nil {
		var err error
		encoded, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encodage de la requête : %w", err)
		}
	}

	usedToken := client.currentAccessToken()
	response, err := client.send(method, path, encoded, usedToken)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusUnauthorized && usedToken != "" && !strings.HasPrefix(path, "/auth/") {
		if client.refresh(usedToken) == nil {
			response.Body.Close()
			response, err = client.send(method, path, encoded, client.currentAccessToken())
			if err != nil {
				return err
			}
		}
	}
	defer response.Body.Close()
	return decodeResponse(response, method, path, destination)
}

func (client *Client) send(method, path string, encoded []byte, accessToken string) (*http.Response, error) {
	var requestBody io.Reader
	if encoded != nil {
		requestBody = bytes.NewReader(encoded)
	}

	request, err := http.NewRequest(method, client.baseURL+path, requestBody)
	if err != nil {
		return nil, fmt.Errorf("création de la requête : %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("API injoignable : %w", err)
	}
	return response, nil
}

func (client *Client) refresh(expiredToken string) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.accessToken != expiredToken {
		return nil
	}
	if client.refreshToken == "" {
		return &APIError{Status: http.StatusUnauthorized, Message: "session expirée"}
	}

	encoded, err := json.Marshal(map[string]string{"refresh_token": client.refreshToken})
	if err != nil {
		return fmt.Errorf("encodage de la requête : %w", err)
	}
	response, err := client.send(http.MethodPost, "/auth/refresh", encoded, "")
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var tokens tokensResponse
	if err := decodeResponse(response, http.MethodPost, "/auth/refresh", &tokens); err != nil {
		return err
	}
	client.accessToken = tokens.AccessToken
	client.refreshToken = tokens.RefreshToken
	return nil
}

func decodeResponse(response *http.Response, method, path string, destination any) error {
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var errorBody struct {
			Error string `json:"error"`
		}
		if json.NewDecoder(response.Body).Decode(&errorBody) != nil || errorBody.Error == "" {
			errorBody.Error = fmt.Sprintf("%s (%s %s)", response.Status, method, path)
		}
		return &APIError{Status: response.StatusCode, Message: errorBody.Error}
	}

	if destination == nil {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		return fmt.Errorf("décodage de la réponse : %w", err)
	}
	return nil
}
