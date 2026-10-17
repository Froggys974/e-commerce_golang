package apiclient

import (
	"net/http"
	"net/url"

	"github.com/Froggys974/e-commerce_golang/internal/product"
)

func (client *Client) SearchProducts(name string) ([]product.Product, error) {
	query := url.Values{}
	if name != "" {
		query.Set("name", name)
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
