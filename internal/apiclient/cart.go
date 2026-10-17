package apiclient

import (
	"net/http"
	"net/url"

	"github.com/Froggys974/e-commerce_golang/internal/cart"
)

type setCartItemRequest struct {
	Quantity int `json:"quantity"`
}

func (client *Client) Cart() (cart.Cart, error) {
	var current cart.Cart
	err := client.do(http.MethodGet, "/cart", nil, &current)
	return current, err
}

func (client *Client) SetCartItem(productRef string, quantity int) (cart.Cart, error) {
	var current cart.Cart
	err := client.do(http.MethodPut, "/cart/items/"+url.PathEscape(productRef), setCartItemRequest{Quantity: quantity}, &current)
	return current, err
}

func (client *Client) RemoveCartItem(productRef string) (cart.Cart, error) {
	var current cart.Cart
	err := client.do(http.MethodDelete, "/cart/items/"+url.PathEscape(productRef), nil, &current)
	return current, err
}
