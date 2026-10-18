package admin

import (
	"github.com/Froggys974/e-commerce_golang/internal/order"
	"github.com/Froggys974/e-commerce_golang/internal/price"
)

func formatPrice(cents int64) string {
	return price.Format(cents)
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
