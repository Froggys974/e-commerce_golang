package client

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/cart"
)

func newCartTable() table.Model {
	return table.New(
		table.WithColumns([]table.Column{
			{Title: "Produit", Width: 30},
			{Title: "Qté", Width: 5},
			{Title: "Prix unitaire", Width: 14},
			{Title: "Total", Width: 14},
		}),
		table.WithHeight(10),
		table.WithWidth(70),
		table.WithFocused(true),
	)
}

func loadCartCmd(api *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		current, err := api.Cart()
		return cartUpdatedMsg{cart: current, err: err}
	}
}

// quantite zero retire le produit
func setQuantityCmd(api *apiclient.Client, productRef string, quantity int) tea.Cmd {
	return func() tea.Msg {
		var updated cart.Cart
		var err error
		if quantity <= 0 {
			updated, err = api.RemoveCartItem(productRef)
		} else {
			updated, err = api.SetCartItem(productRef, quantity)
		}
		return cartUpdatedMsg{cart: updated, err: err}
	}
}

func (model catalog) updateCart(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		model.showCart = false
		model.status = ""
		return model, nil
	case "o":
		return model.openOrders()
	case "up", "down":
		var cmd tea.Cmd
		model.cartTable, cmd = model.cartTable.Update(msg)
		return model, cmd
	}

	if len(model.cart.Items) == 0 {
		return model, nil
	}
	line := model.cart.Items[model.cartTable.Cursor()]

	switch msg.String() {
	case "+":
		return model, setQuantityCmd(model.api, line.ProductRef, line.Quantity+1)
	case "-":
		return model, setQuantityCmd(model.api, line.ProductRef, line.Quantity-1)
	case "delete", "backspace":
		return model, setQuantityCmd(model.api, line.ProductRef, 0)
	case "p":
		model.err = nil
		model.status = ""
		model.checkout = newCheckoutForm()
		return model, model.checkout.Init()
	}
	return model, nil
}

func (model catalog) cartView() string {
	if model.checkout != nil {
		return model.checkoutView()
	}

	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Panier " + model.cart.Ref))
	builder.WriteString("\n\n")

	if len(model.cart.Items) == 0 {
		builder.WriteString("Ton panier est vide\n\n")
	} else {
		builder.WriteString(model.cartTable.View())
		builder.WriteString("\n\n")
	}

	fmt.Fprintf(&builder, "Total TTC : %s (livraison gratuite)\n\n", formatPrice(model.cart.TotalTTCCents))
	builder.WriteString(helpStyle.Render("↑/↓ naviguer • + / - quantité • suppr retirer • p payer • o commandes • échap catalogue"))
	builder.WriteString(model.feedbackView())
	return builder.String()
}

func cartRows(current cart.Cart) []table.Row {
	rows := make([]table.Row, 0, len(current.Items))
	for _, line := range current.Items {
		rows = append(rows, table.Row{
			line.Name,
			strconv.Itoa(line.Quantity),
			formatPrice(line.UnitPriceTTCCents),
			formatPrice(line.LineTotalCents),
		})
	}
	return rows
}
