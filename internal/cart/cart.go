package cart

import (
	"errors"
	"fmt"
)

type Item struct {
	ProductRef        string `json:"product_ref"`
	Name              string `json:"name"`
	Quantity          int    `json:"quantity"`
	UnitPriceTTCCents int64  `json:"unit_price_ttc_cents"`
	LineTotalCents    int64  `json:"line_total_cents"`
}

type Cart struct {
	Ref           string `json:"ref"`
	Items         []Item `json:"items"`
	TotalTTCCents int64  `json:"total_ttc_cents"`
}

var (
	ErrRefTaken        = errors.New("référence panier déjà utilisée")
	ErrProductNotFound = errors.New("produit introuvable")
	ErrItemNotFound    = errors.New("produit absent du panier")
)

type ValidationError struct {
	Field   string
	Message string
}

func (err ValidationError) Error() string {
	return fmt.Sprintf("%s : %s", err.Field, err.Message)
}
