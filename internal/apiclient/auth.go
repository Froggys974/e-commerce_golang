package apiclient

import (
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (client *Client) Login(email, password string) error {
	var tokens tokensResponse
	err := client.do(http.MethodPost, "/auth/login", loginRequest{Email: email, Password: password}, &tokens)
	if err != nil {
		return err
	}
	client.setTokens(tokens)
	return nil
}

func (client *Client) Me() (user.User, error) {
	var account user.User
	err := client.do(http.MethodGet, "/me", nil, &account)
	return account, err
}

func (client *Client) Register(email, password string) error {
	return client.do(http.MethodPost, "/auth/register", loginRequest{Email: email, Password: password}, nil)
}

type confirmRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (client *Client) Confirm(email, code string) error {
	return client.do(http.MethodPost, "/auth/confirm", confirmRequest{Email: email, Code: code}, nil)
}

type emailRequest struct {
	Email string `json:"email"`
}

func (client *Client) ResendConfirmCode(email string) error {
	return client.do(http.MethodPost, "/auth/confirm/resend", emailRequest{Email: email}, nil)
}

func (client *Client) ForgotPassword(email string) error {
	return client.do(http.MethodPost, "/auth/forgot", emailRequest{Email: email}, nil)
}

type resetRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (client *Client) ResetPassword(email, code, newPassword string) error {
	return client.do(http.MethodPost, "/auth/reset", resetRequest{Email: email, Code: code, NewPassword: newPassword}, nil)
}
