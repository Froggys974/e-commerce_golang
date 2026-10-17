package admin

import (
	"errors"
	"net/http"
	"time"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/product"
	"github.com/Froggys974/e-commerce_golang/internal/ref"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type orderDraft struct {
	customer user.User
	lines    []orderLine
}

type orderLine struct {
	product  product.Product
	quantity int
}

type ordersSource interface {
	List() ([]apiclient.Order, error)
	UpdateStatus(orderRef string, status apiclient.OrderStatus, reason string) error
	Create(draft orderDraft) error
}

// --- API ---

type apiOrders struct {
	api *apiclient.Client
}

func (source apiOrders) List() ([]apiclient.Order, error) {
	return source.api.ListAllOrders()
}

func (source apiOrders) UpdateStatus(orderRef string, status apiclient.OrderStatus, reason string) error {
	return source.api.UpdateOrderStatus(orderRef, status, reason)
}

func (source apiOrders) Create(draft orderDraft) error {
	items := make([]apiclient.NewOrderItem, 0, len(draft.lines))
	for _, line := range draft.lines {
		items = append(items, apiclient.NewOrderItem{Ref: line.product.Ref, Quantity: line.quantity})
	}
	return source.api.CreateOrder(draft.customer.ID, items)
}

// route absente côté API = 404 (ou 405 si la route existe avec une autre méthode)
func isMissingRoute(err error) bool {
	var apiError *apiclient.APIError
	return errors.As(err, &apiError) &&
		(apiError.Status == http.StatusNotFound || apiError.Status == http.StatusMethodNotAllowed)
}

// --- PLACEHOLDER ---

type fakeOrders struct {
	orders []apiclient.Order
}

func newFakeOrders() *fakeOrders {
	now := time.Now()
	return &fakeOrders{orders: []apiclient.Order{
		{
			Ref: "CMD-DEMO01", UserID: 2, UserEmail: "alice@example.com", Status: apiclient.OrderPaid,
			TotalTTCCents: 10798, CreatedAt: now.Add(-48 * time.Hour),
			Items: []apiclient.OrderItem{
				{ProductRef: "PDT-DEMO01", ProductName: "Clavier mécanique", Quantity: 1, UnitPriceTTCCents: 10798},
			},
		},
		{
			Ref: "CMD-DEMO02", UserID: 3, UserEmail: "bob@example.com", Status: apiclient.OrderShipping,
			TotalTTCCents: 4536, CreatedAt: now.Add(-24 * time.Hour),
			Items: []apiclient.OrderItem{
				{ProductRef: "PDT-DEMO02", ProductName: "Café en grains", Quantity: 2, UnitPriceTTCCents: 2268},
			},
		},
		{
			Ref: "CMD-DEMO03", UserID: 2, UserEmail: "alice@example.com", Status: apiclient.OrderCancelled,
			CancelReason: "rupture de stock", TotalTTCCents: 900, CreatedAt: now.Add(-2 * time.Hour),
			Items: []apiclient.OrderItem{
				{ProductRef: "PDT-DEMO03", ProductName: "Thé vert", Quantity: 1, UnitPriceTTCCents: 900},
			},
		},
	}}
}

func (source *fakeOrders) List() ([]apiclient.Order, error) {
	return append([]apiclient.Order(nil), source.orders...), nil
}

func (source *fakeOrders) UpdateStatus(orderRef string, status apiclient.OrderStatus, reason string) error {
	for index := range source.orders {
		if source.orders[index].Ref == orderRef {
			source.orders[index].Status = status
			source.orders[index].CancelReason = reason
			return nil
		}
	}
	return errors.New("commande introuvable")
}

func (source *fakeOrders) Create(draft orderDraft) error {
	order := apiclient.Order{
		Ref:       ref.New(ref.Order),
		UserID:    draft.customer.ID,
		UserEmail: draft.customer.Email,
		Status:    apiclient.OrderPending,
		CreatedAt: time.Now(),
	}
	for _, line := range draft.lines {
		order.Items = append(order.Items, apiclient.OrderItem{
			ProductRef:        line.product.Ref,
			ProductName:       line.product.Name,
			Quantity:          line.quantity,
			UnitPriceTTCCents: line.product.PriceTTCCents,
		})
		order.TotalTTCCents += line.product.PriceTTCCents * int64(line.quantity)
	}
	source.orders = append(source.orders, order)
	return nil
}
