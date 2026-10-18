package apiclient

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/Froggys974/e-commerce_golang/internal/product"
)

func (client *Client) SearchProducts(filters product.SearchFilters) ([]product.Product, error) {
	query := url.Values{}
	textFilters := map[string]string{
		"name":        filters.Name,
		"description": filters.Description,
		"category":    filters.Category,
	}
	for param, value := range textFilters {
		if value != "" {
			query.Set(param, value)
		}
	}
	priceFilters := map[string]*int64{
		"min_price_ht":  filters.MinPriceHT,
		"max_price_ht":  filters.MaxPriceHT,
		"min_price_ttc": filters.MinPriceTTC,
		"max_price_ttc": filters.MaxPriceTTC,
	}
	for param, cents := range priceFilters {
		if cents != nil {
			query.Set(param, strconv.FormatInt(*cents, 10))
		}
	}

	var products []product.Product
	err := client.do(http.MethodGet, "/products?"+query.Encode(), nil, &products)
	return products, err
}

type CreateProductInput struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	PriceHTCents int64  `json:"price_ht_cents"`
	VATPercent   int    `json:"vat_percent"`
}

func (client *Client) CreateProduct(input CreateProductInput) (product.Product, error) {
	var created product.Product
	err := client.do(http.MethodPost, "/admin/products", input, &created)
	return created, err
}
