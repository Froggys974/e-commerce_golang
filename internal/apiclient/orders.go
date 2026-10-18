package apiclient

import (
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/order"
)

type CheckoutRequest struct {
	CardNumber string `json:"card_number"`
	ExpMonth   int    `json:"exp_month"`
	ExpYear    int    `json:"exp_year"`
	CVC        string `json:"cvc"`
}

func (client *Client) Checkout(card CheckoutRequest) (order.Order, error) {
	var created order.Order
	err := client.do(http.MethodPost, "/cart/checkout", card, &created)
	return created, err
}

func (client *Client) MyOrders() ([]order.Order, error) {
	var orders []order.Order
	err := client.do(http.MethodGet, "/orders", nil, &orders)
	return orders, err
}
