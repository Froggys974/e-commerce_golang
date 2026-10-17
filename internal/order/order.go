package order

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusPaid      Status = "paid"
	StatusShipping  Status = "shipping"
	StatusDelivered Status = "delivered"
	StatusCancelled Status = "cancelled"
)

var Statuses = []Status{StatusPending, StatusPaid, StatusShipping, StatusDelivered, StatusCancelled}

func (status Status) Valid() bool {
	return slices.Contains(Statuses, status)
}

type Item struct {
	ProductRef        string `json:"product_ref"`
	ProductName       string `json:"product_name"`
	Quantity          int    `json:"quantity"`
	UnitPriceTTCCents int64  `json:"unit_price_ttc_cents"`
}

type Order struct {
	Ref           string    `json:"ref"`
	UserID        int64     `json:"user_id"`
	UserEmail     string    `json:"user_email"`
	Status        Status    `json:"status"`
	CancelReason  string    `json:"cancel_reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	Items         []Item    `json:"items"`
	TotalTTCCents int64     `json:"total_ttc_cents"`
}

type ItemInput struct {
	ProductRef string `json:"ref"`
	Quantity   int    `json:"quantity"`
}

type Card struct {
	Number   string
	ExpMonth int
	ExpYear  int
	CVC      string
}

var (
	ErrRefTaken        = errors.New("référence commande déjà utilisée")
	ErrEmptyCart       = errors.New("panier vide")
	ErrNotFound        = errors.New("commande introuvable")
	ErrProductNotFound = errors.New("produit introuvable")
	ErrUserNotFound    = errors.New("utilisateur introuvable")
)

type ValidationError struct {
	Field   string
	Message string
}

func (err ValidationError) Error() string {
	return fmt.Sprintf("%s : %s", err.Field, err.Message)
}
