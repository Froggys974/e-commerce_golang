package client

import (
	"fmt"
)

func formatPrice(cents int64) string {
	// 1999 -> 19,99 €
	return fmt.Sprintf("%d,%02d €", cents/100, cents%100)
}
