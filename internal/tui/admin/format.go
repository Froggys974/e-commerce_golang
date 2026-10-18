package admin

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Froggys974/e-commerce_golang/internal/order"
)

var errEurosFormat = errors.New("format attendu : 19,99")

// 1999 -> "19,99 €"
func formatPrice(cents int64) string {
	return fmt.Sprintf("%d,%02d €", cents/100, cents%100)
}

// "19,99" ou "19.99" ou "19" -> 1999
func parseEuros(text string) (int64, error) {
	text = strings.ReplaceAll(strings.TrimSpace(text), ",", ".")
	euros, cents, hasCents := strings.Cut(text, ".")
	if !isDigits(euros) || (hasCents && (!isDigits(cents) || len(cents) > 2)) {
		return 0, errEurosFormat
	}
	if len(cents) == 1 {
		cents += "0"
	}

	eurosValue, err := strconv.ParseInt(euros, 10, 64)
	// borne pour que eurosValue fois 100 plus 99 tienne dans un int64
	if err != nil || eurosValue > (math.MaxInt64-99)/100 {
		return 0, errEurosFormat
	}
	var centsValue int64
	if hasCents {
		// deja verifie par isDigits
		centsValue, _ = strconv.ParseInt(cents, 10, 64)
	}
	return eurosValue*100 + centsValue, nil
}

func isDigits(text string) bool {
	if text == "" {
		return false
	}
	for _, character := range text {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

var statusLabels = map[order.Status]string{
	order.StatusPending:   "En attente",
	order.StatusPaid:      "Payé",
	order.StatusShipping:  "En cours de livraison",
	order.StatusDelivered: "Livré",
	order.StatusCancelled: "Annulé",
}

func statusLabel(status order.Status) string {
	if label, ok := statusLabels[status]; ok {
		return label
	}
	return string(status)
}
