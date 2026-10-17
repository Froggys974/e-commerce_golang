package client

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/order"
)

type ordersLoadedMsg struct {
	orders []order.Order
	err    error
}

func newOrdersTable() table.Model {
	return table.New(
		table.WithColumns([]table.Column{
			{Title: "Référence", Width: 12},
			{Title: "Date", Width: 17},
			{Title: "Statut", Width: 22},
			{Title: "Articles", Width: 9},
			{Title: "Total", Width: 12},
		}),
		table.WithHeight(10),
		table.WithWidth(80),
		table.WithFocused(true),
	)
}

func loadOrdersCmd(api *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		orders, err := api.MyOrders()
		return ordersLoadedMsg{orders: orders, err: err}
	}
}

func (model catalog) openOrders() (catalog, tea.Cmd) {
	model.showCart = false
	model.showOrders = true
	model.selectedOrder = nil
	model.status = ""
	model.err = nil
	return model, loadOrdersCmd(model.api)
}

func (model catalog) applyOrders(msg ordersLoadedMsg) catalog {
	model.err = msg.err
	if msg.err != nil {
		return model
	}
	model.orders = msg.orders
	model.ordersTable.SetRows(orderRows(msg.orders))
	model.ordersTable.SetCursor(0)
	return model
}

func (model catalog) updateOrders(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if model.selectedOrder != nil {
		if msg.String() == "esc" {
			model.selectedOrder = nil
		}
		return model, nil
	}

	switch msg.String() {
	case "esc":
		model.showOrders = false
		return model, nil
	case "enter":
		if len(model.orders) > 0 {
			model.selectedOrder = &model.orders[model.ordersTable.Cursor()]
		}
		return model, nil
	case "up", "down":
		var cmd tea.Cmd
		model.ordersTable, cmd = model.ordersTable.Update(msg)
		return model, cmd
	}
	return model, nil
}

func (model catalog) ordersView() string {
	if model.selectedOrder != nil {
		return orderDetailView(*model.selectedOrder)
	}

	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Mes commandes"))
	builder.WriteString("\n\n")
	if len(model.orders) == 0 {
		builder.WriteString("Tu n'as pas encore de commande\n\n")
	} else {
		builder.WriteString(model.ordersTable.View())
		builder.WriteString("\n\n")
	}
	builder.WriteString(helpStyle.Render("↑/↓ naviguer • entrée détail • échap catalogue"))
	builder.WriteString(model.feedbackView())
	return builder.String()
}

func orderDetailView(current order.Order) string {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Commande " + current.Ref))
	builder.WriteString("\n\n")
	fmt.Fprintf(&builder, "Date     %s\n", current.CreatedAt.Local().Format("02/01/2006 15:04"))
	fmt.Fprintf(&builder, "Statut   %s\n", statusLabel(current.Status))
	if current.CancelReason != "" {
		fmt.Fprintf(&builder, "Raison   %s\n", current.CancelReason)
	}
	builder.WriteString("\n")

	for _, item := range current.Items {
		fmt.Fprintf(&builder, "%-30s x%-3d %s\n", item.ProductName, item.Quantity, formatPrice(item.UnitPriceTTCCents*int64(item.Quantity)))
	}
	fmt.Fprintf(&builder, "\nTotal TTC : %s\n\n", formatPrice(current.TotalTTCCents))
	builder.WriteString(helpStyle.Render("échap retour"))
	return builder.String()
}

func orderRows(orders []order.Order) []table.Row {
	rows := make([]table.Row, 0, len(orders))
	for _, current := range orders {
		count := 0
		for _, item := range current.Items {
			count += item.Quantity
		}
		rows = append(rows, table.Row{
			current.Ref,
			current.CreatedAt.Local().Format("02/01/2006 15:04"),
			statusLabel(current.Status),
			strconv.Itoa(count),
			formatPrice(current.TotalTTCCents),
		})
	}
	return rows
}

func statusLabel(status order.Status) string {
	switch status {
	case order.StatusPending:
		return "en attente"
	case order.StatusPaid:
		return "payée"
	case order.StatusShipping:
		return "en cours de livraison"
	case order.StatusDelivered:
		return "livrée"
	case order.StatusCancelled:
		return "annulée"
	}
	return string(status)
}
