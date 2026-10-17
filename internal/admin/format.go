package admin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
)

// 1999 -> "19,99 €"
func formatPrice(cents int64) string {
	return fmt.Sprintf("%d,%02d €", cents/100, cents%100)
}

// "19,99" ou "19.99" ou "19" -> 1999
func parseEuros(text string) (int64, error) {
	text = strings.ReplaceAll(strings.TrimSpace(text), ",", ".")
	euros, cents, hasCents := strings.Cut(text, ".")
	if euros == "" || (hasCents && (len(cents) == 0 || len(cents) > 2)) {
		return 0, errors.New("format attendu : 19,99")
	}
	if len(cents) == 1 {
		cents += "0"
	}

	eurosValue, err := strconv.ParseInt(euros, 10, 64)
	if err != nil || eurosValue < 0 {
		return 0, errors.New("format attendu : 19,99")
	}
	var centsValue int64
	if hasCents {
		centsValue, err = strconv.ParseInt(cents, 10, 64)
		if err != nil || centsValue < 0 {
			return 0, errors.New("format attendu : 19,99")
		}
	}
	return eurosValue*100 + centsValue, nil
}

var statusLabels = map[apiclient.OrderStatus]string{
	apiclient.OrderPending:   "En attente",
	apiclient.OrderPaid:      "Payé",
	apiclient.OrderShipping:  "En cours de livraison",
	apiclient.OrderDelivered: "Livré",
	apiclient.OrderCancelled: "Annulé",
}

func statusLabel(status apiclient.OrderStatus) string {
	if label, ok := statusLabels[status]; ok {
		return label
	}
	return string(status)
}
