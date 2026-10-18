package apiclient

import (
	"net/http"
	"net/url"

	"github.com/Froggys974/e-commerce_golang/internal/order"
)

type updateOrderStatusRequest struct {
	Status order.Status `json:"status"`
	Reason string       `json:"reason,omitempty"`
}

type createOrderRequest struct {
	UserID int64             `json:"user_id"`
	Items  []order.ItemInput `json:"items"`
}

func (client *Client) ListAllOrders() ([]order.Order, error) {
	var orders []order.Order
	err := client.do(http.MethodGet, "/admin/orders", nil, &orders)
	return orders, err
}

func (client *Client) UpdateOrderStatus(orderRef string, status order.Status, reason string) error {
	return client.do(http.MethodPatch, "/admin/orders/"+url.PathEscape(orderRef), updateOrderStatusRequest{Status: status, Reason: reason}, nil)
}

func (client *Client) CreateOrder(userID int64, items []order.ItemInput) error {
	return client.do(http.MethodPost, "/admin/orders", createOrderRequest{UserID: userID, Items: items}, nil)
}
