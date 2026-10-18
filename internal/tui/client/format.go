package client

import "github.com/Froggys974/e-commerce_golang/internal/price"

// 1999 -> 19,99 €
func formatPrice(cents int64) string {
	return price.Format(cents)
}
