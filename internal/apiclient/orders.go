package apiclient

import (
	"net/http"
	"time"
)

// PLACEHOLDER : routes et JSON
//	GET   /admin/orders
//	PATCH /admin/orders/{ref}
//	POST  /admin/orders

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPaid      OrderStatus = "paid"
	OrderShipping  OrderStatus = "shipping"
	OrderDelivered OrderStatus = "delivered"
	OrderCancelled OrderStatus = "cancelled"
)

var OrderStatuses = []OrderStatus{OrderPending, OrderPaid, OrderShipping, OrderDelivered, OrderCancelled}

type OrderItem struct {
	ProductRef        string `json:"product_ref"`
	ProductName       string `json:"product_name"`
	Quantity          int    `json:"quantity"`
	UnitPriceTTCCents int64  `json:"unit_price_ttc_cents"`
}

type Order struct {
	Ref           string      `json:"ref"`
	UserID        int64       `json:"user_id"`
	UserEmail     string      `json:"user_email"`
	Status        OrderStatus `json:"status"`
	CancelReason  string      `json:"cancel_reason,omitempty"`
	TotalTTCCents int64       `json:"total_ttc_cents"`
	Items         []OrderItem `json:"items"`
	CreatedAt     time.Time   `json:"created_at"`
}

type updateOrderStatusRequest struct {
	Status OrderStatus `json:"status"`
	Reason string      `json:"reason,omitempty"`
}

type NewOrderItem struct {
	Ref      string `json:"ref"`
	Quantity int    `json:"quantity"`
}

type createOrderRequest struct {
	UserID int64          `json:"user_id"`
	Items  []NewOrderItem `json:"items"`
}

func (client *Client) ListAllOrders() ([]Order, error) {
	var orders []Order
	err := client.do(http.MethodGet, "/admin/orders", nil, &orders)
	return orders, err
}

func (client *Client) UpdateOrderStatus(ref string, status OrderStatus, reason string) error {
	return client.do(http.MethodPatch, "/admin/orders/"+ref, updateOrderStatusRequest{Status: status, Reason: reason}, nil)
}

func (client *Client) CreateOrder(userID int64, items []NewOrderItem) error {
	return client.do(http.MethodPost, "/admin/orders", createOrderRequest{UserID: userID, Items: items}, nil)
}
