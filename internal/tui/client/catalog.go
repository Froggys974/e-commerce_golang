package client

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/cart"
	"github.com/Froggys974/e-commerce_golang/internal/order"
	"github.com/Froggys974/e-commerce_golang/internal/product"
)

type searchResultMsg struct {
	query    string
	products []product.Product
	err      error
}

type cartUpdatedMsg struct {
	productName string
	cart        cart.Cart
	err         error
}

type catalog struct {
	api       *apiclient.Client
	search    textinput.Model
	table     table.Model
	products  []product.Product
	selected  *product.Product
	showCart  bool
	cart      cart.Cart
	cartTable table.Model
	checkout  *huh.Form

	showOrders    bool
	orders        []order.Order
	ordersTable   table.Model
	selectedOrder *order.Order
	status        string
	err           error
}

func newCatalog(api *apiclient.Client) catalog {
	search := textinput.New()
	search.Placeholder = "rechercher par nom..."
	search.SetWidth(40)
	search.Focus()

	productTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "Réf", Width: 11},
			{Title: "Nom", Width: 30},
			{Title: "Catégorie", Width: 15},
			{Title: "Prix TTC", Width: 14},
		}),
		table.WithHeight(10),
		table.WithWidth(80),
		table.WithFocused(true),
	)

	return catalog{
		api:         api,
		search:      search,
		table:       productTable,
		cartTable:   newCartTable(),
		ordersTable: newOrdersTable(),
	}
}

func (model catalog) Init() tea.Cmd {
	return searchCmd(model.api, "")
}

func searchCmd(api *apiclient.Client, query string) tea.Cmd {
	return func() tea.Msg {
		products, err := api.SearchProducts(query)
		return searchResultMsg{query: query, products: products, err: err}
	}
}

// quantite actuelle plus un
func addToCartCmd(api *apiclient.Client, item product.Product) tea.Cmd {
	return func() tea.Msg {
		current, err := api.Cart()
		if err != nil {
			return cartUpdatedMsg{err: err}
		}

		quantity := 1
		for _, line := range current.Items {
			if line.ProductRef == item.Ref {
				quantity = line.Quantity + 1
			}
		}

		updated, err := api.SetCartItem(item.Ref, quantity)
		return cartUpdatedMsg{productName: item.Name, cart: updated, err: err}
	}
}

func (model catalog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if done, ok := msg.(checkoutDoneMsg); ok {
		return model.applyCheckout(done), nil
	}
	if loaded, ok := msg.(ordersLoadedMsg); ok {
		return model.applyOrders(loaded), nil
	}
	if model.checkout != nil {
		return model.updateCheckout(msg)
	}

	switch msg := msg.(type) {
	case searchResultMsg:
		// ignorer ancienne reponse qui arrive en retard
		if msg.query != model.search.Value() {
			return model, nil
		}
		model.err = msg.err
		model.products = msg.products
		model.table.SetRows(toRows(msg.products))
		model.table.SetCursor(0)
		return model, nil

	case cartUpdatedMsg:
		model.err = msg.err
		model.status = ""
		if msg.err != nil {
			return model, nil
		}
		model.cart = msg.cart
		model.cartTable.SetRows(cartRows(msg.cart))
		if model.cartTable.Cursor() >= len(msg.cart.Items) {
			model.cartTable.SetCursor(max(len(msg.cart.Items)-1, 0))
		}
		if msg.productName != "" {
			model.status = fmt.Sprintf("✓ %s ajouté au panier • %d article(s) • total %s",
				msg.productName, countItems(msg.cart), formatPrice(msg.cart.TotalTTCCents))
		}
		return model, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return model, tea.Quit
		}
		if model.showOrders {
			return model.updateOrders(msg)
		}
		if model.showCart {
			return model.updateCart(msg)
		}
		if model.selected != nil {
			return model.updateDetail(msg)
		}
		return model.updateList(msg)
	}

	return model, nil
}

func (model catalog) updateList(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return model, tea.Quit
	case "enter":
		if len(model.products) > 0 {
			model.selected = &model.products[model.table.Cursor()]
		}
		return model, nil
	case "up", "down":
		var cmd tea.Cmd
		model.table, cmd = model.table.Update(msg)
		return model, cmd
	case "+":
		if len(model.products) > 0 {
			return model, addToCartCmd(model.api, model.products[model.table.Cursor()])
		}
		return model, nil
	case "tab":
		model.showCart = true
		model.status = ""
		return model, loadCartCmd(model.api)
	case "ctrl+o":
		return model.openOrders()
	}

	previousQuery := model.search.Value()
	var cmd tea.Cmd
	model.search, cmd = model.search.Update(msg)
	if model.search.Value() != previousQuery {
		return model, tea.Batch(cmd, searchCmd(model.api, model.search.Value()))
	}
	return model, cmd
}

func (model catalog) updateDetail(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		model.selected = nil
	case "a", "+":
		return model, addToCartCmd(model.api, *model.selected)
	}
	return model, nil
}

func (model catalog) View() tea.View {
	var screen string
	switch {
	case model.showOrders:
		screen = model.ordersView()
	case model.showCart:
		screen = model.cartView()
	case model.selected != nil:
		screen = detailView(*model.selected) + model.feedbackView()
	default:
		screen = model.listView()
	}
	view := tea.NewView(screen)
	view.AltScreen = true
	return view
}

func (model catalog) listView() string {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Catalogue"))
	builder.WriteString("\n\n")
	builder.WriteString(model.search.View())
	builder.WriteString("\n\n")
	builder.WriteString(model.table.View())
	builder.WriteString("\n\n")
	builder.WriteString(helpStyle.Render(fmt.Sprintf("%d produit(s) • ↑/↓ naviguer • + ajouter au panier • tab panier • ctrl+o commandes • entrée détail • échap quitter", len(model.products))))
	builder.WriteString(model.feedbackView())
	return builder.String()
}

func (model catalog) feedbackView() string {
	switch {
	case model.err != nil:
		return "\n\n" + errorStyle.Render("✗ "+model.err.Error())
	case model.status != "":
		return "\n\n" + successStyle.Render(model.status)
	}
	return ""
}

func countItems(current cart.Cart) int {
	count := 0
	for _, line := range current.Items {
		count += line.Quantity
	}
	return count
}

func detailView(item product.Product) string {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render(item.Name))
	builder.WriteString("\n\n")
	builder.WriteString("Référence   ")
	builder.WriteString(item.Ref)
	builder.WriteString("\n")
	builder.WriteString("Catégorie   ")
	builder.WriteString(item.Category)
	builder.WriteString("\n")
	builder.WriteString("Prix HT     ")
	builder.WriteString(formatPrice(item.PriceHTCents))
	builder.WriteString("\n")
	fmt.Fprintf(&builder, "TVA         %d %%\n", item.VATPercent)
	builder.WriteString("Prix TTC    ")
	builder.WriteString(formatPrice(item.PriceTTCCents))
	builder.WriteString("\n\n")
	builder.WriteString(item.Description)
	builder.WriteString("\n\n")
	builder.WriteString(helpStyle.Render("a ajouter au panier • échap retour"))
	return builder.String()
}

func toRows(products []product.Product) []table.Row {
	rows := make([]table.Row, 0, len(products))
	for _, item := range products {
		rows = append(rows, table.Row{item.Ref, item.Name, item.Category, formatPrice(item.PriceTTCCents)})
	}
	return rows
}
