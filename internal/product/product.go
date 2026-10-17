package product

import (
	"errors"
	"fmt"
	"time"
)

type Product struct {
	ID            int64     `json:"id"`
	Ref           string    `json:"ref"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	PriceHTCents  int64     `json:"price_ht_cents"`
	VATPercent    int       `json:"vat_percent"`
	PriceTTCCents int64     `json:"price_ttc_cents"`
	CreatedAt     time.Time `json:"created_at"`
}

// arrondi au centime le plus proche
func PriceTTC(priceHTCents int64, vatPercent int) int64 {
	return (priceHTCents*int64(100+vatPercent) + 50) / 100
}

var (
	ErrRefTaken = errors.New("référence produit déjà utilisée")
	ErrNotFound = errors.New("produit introuvable")
)

type SearchFilters struct {
	Name        string
	Description string
	Category    string
	MinPriceHT  *int64
	MaxPriceHT  *int64
	MinPriceTTC *int64
	MaxPriceTTC *int64
}

type ValidationError struct {
	Field   string
	Message string
}

func (err ValidationError) Error() string {
	return fmt.Sprintf("%s : %s", err.Field, err.Message)
}
