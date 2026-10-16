package apiclient

import (
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (client *Client) Login(email, password string) error {
	var tokens tokensResponse
	err := client.do(http.MethodPost, "/auth/login", loginRequest{Email: email, Password: password}, &tokens)
	if err != nil {
		return err
	}
	client.accessToken = tokens.AccessToken
	client.refreshToken = tokens.RefreshToken
	return nil
}

func (client *Client) Me() (user.User, error) {
	var account user.User
	err := client.do(http.MethodGet, "/me", nil, &account)
	return account, err
}
