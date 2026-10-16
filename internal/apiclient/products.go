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
