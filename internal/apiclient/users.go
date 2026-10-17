package apiclient

import (
	"fmt"
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type userRequest struct {
	Email    string    `json:"email"`
	Password string    `json:"password,omitempty"`
	Role     user.Role `json:"role"`
}

func (client *Client) ListUsers() ([]user.User, error) {
	var users []user.User
	err := client.do(http.MethodGet, "/admin/users", nil, &users)
	return users, err
}

func (client *Client) CreateUser(email, password string, role user.Role) error {
	return client.do(http.MethodPost, "/admin/users", userRequest{Email: email, Password: password, Role: role}, nil)
}

func (client *Client) UpdateUser(id int64, email string, role user.Role) error {
	return client.do(http.MethodPut, fmt.Sprintf("/admin/users/%d", id), userRequest{Email: email, Role: role}, nil)
}

func (client *Client) DeleteUser(id int64) error {
	return client.do(http.MethodDelete, fmt.Sprintf("/admin/users/%d", id), nil, nil)
}

func (client *Client) ConfirmUser(id int64) error {
	return client.do(http.MethodPost, fmt.Sprintf("/admin/users/%d/confirm", id), nil, nil)
}
