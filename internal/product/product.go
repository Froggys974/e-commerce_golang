package product

import (
	"database/sql"
)

type Handler struct {
	database *sql.DB
}

func NewHandler(database *sql.DB) *Handler {
	return &Handler{database: database}
}

type Product struct {
	Ref           string `json:"ref"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	PriceHTCents  int64  `json:"price_ht_cents"`
	VATPercent    int    `json:"vat_percent"`
	PriceTTCCents int64  `json:"price_ttc_cents"`
}

func PriceTTC(priceHTCents int64, vatPercent int) int64 {
	return (priceHTCents*int64(100+vatPercent) + 50) / 100
}
