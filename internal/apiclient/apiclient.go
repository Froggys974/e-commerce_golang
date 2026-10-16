package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL      string
	httpClient   *http.Client
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

// --> Les méthodes client passent ici <--
func (client *Client) do(method, path string, body any, destination any) error {
	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encodage de la requête : %w", err)
		}
		requestBody = bytes.NewReader(encoded)
	}

	request, err := http.NewRequest(method, client.baseURL+path, requestBody)
	if err != nil {
		return fmt.Errorf("création de la requête : %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if client.accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+client.accessToken)
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("API injoignable : %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		var errorBody struct {
			Error string `json:"error"`
		}
		if json.NewDecoder(response.Body).Decode(&errorBody) != nil || errorBody.Error == "" {
			errorBody.Error = response.Status
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
