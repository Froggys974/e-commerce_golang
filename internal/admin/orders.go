package admin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/product"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type ordersLoadedMsg struct {
	orders []apiclient.Order
	err    error
}

type orderFormDataMsg struct {
	users    []user.User
	products []product.Product
	err      error
}

type formKind int

const (
	formNone formKind = iota
	formStatus
	formOrderPick
	formOrderQuantities
)

type statusFields struct {
	orderRef string
	status   apiclient.OrderStatus
	reason   string
}

type newOrderFields struct {
	customerID  int64
	productRefs []string
	quantities  []string
}

type ordersScreen struct {
	api      *apiclient.Client
	source   ordersSource
	demo     bool // a changer apres remplacer placeholder
	table    table.Model
	orders   []apiclient.Order
	detail   *apiclient.Order
	feedback feedback

	form      *huh.Form
	formKind  formKind
	status    *statusFields
	newOrder  *newOrderFields
	customers []user.User
	catalog   []product.Product
}

func newOrdersScreen(api *apiclient.Client) ordersScreen {
	ordersTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "Réf", Width: 11},
			{Title: "Client", Width: 26},
			{Title: "Statut", Width: 22},
			{Title: "Total TTC", Width: 12},
			{Title: "Date", Width: 16},
		}),
		table.WithHeight(12),
		table.WithWidth(97),
		table.WithFocused(true),
	)
	return ordersScreen{api: api, source: apiOrders{api: api}, table: ordersTable}
}

func (m ordersScreen) init() tea.Cmd {
	return loadOrdersCmd(m.source)
}

func loadOrdersCmd(source ordersSource) tea.Cmd {
	return func() tea.Msg {
		orders, err := source.List()
		return ordersLoadedMsg{orders: orders, err: err}
	}
}

func loadOrderFormDataCmd(api *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		users, err := api.ListUsers()
		if err != nil {
			return orderFormDataMsg{err: err}
		}
		products, err := api.SearchProducts("")
		return orderFormDataMsg{users: users, products: products, err: err}
	}
}

func (m ordersScreen) update(msg tea.Msg) (ordersScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case ordersLoadedMsg:
		if isMissingRoute(msg.err) && !m.demo {
			m.demo = true
			m.source = newFakeOrders()
			return m, loadOrdersCmd(m.source)
		}
		if msg.err != nil {
			m.feedback = feedback{text: msg.err.Error(), isError: true}
			return m, nil
		}
		m.orders = msg.orders
		m.table.SetRows(orderRows(msg.orders))
		if m.table.Cursor() >= len(msg.orders) {
			m.table.SetCursor(0)
		}
		return m, nil

	case orderFormDataMsg:
		if msg.err != nil {
			m.feedback = feedback{text: msg.err.Error(), isError: true}
			return m, nil
		}
		return m.openOrderPickForm(msg.users, msg.products)

	case actionDoneMsg:
		m.feedback = feedbackFrom(msg)
		return m, loadOrdersCmd(m.source)
	}

	if m.form != nil {
		return m.updateForm(msg)
	}

	keyPress, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if m.detail != nil {
		if keyPress.String() == "esc" {
			m.detail = nil
		}
		return m, nil
	}

	switch keyPress.String() {
	case "esc":
		return m, openScreen(screenMenu)
	case "up", "down":
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(keyPress)
		return m, cmd
	case "n":
		m.feedback = feedback{text: "chargement des clients et produits..."}
		return m, loadOrderFormDataCmd(m.api)
	}

	if len(m.orders) == 0 {
		return m, nil
	}
	selected := m.orders[m.table.Cursor()]
	switch keyPress.String() {
	case "enter":
		m.detail = &selected
	case "s":
		return m.openStatusForm(selected)
	}
	return m, nil
}

func (m ordersScreen) openStatusForm(order apiclient.Order) (ordersScreen, tea.Cmd) {
	fields := &statusFields{orderRef: order.Ref, status: order.Status}
	m.status = fields

	options := make([]huh.Option[apiclient.OrderStatus], 0, len(apiclient.OrderStatuses))
	for _, status := range apiclient.OrderStatuses {
		options = append(options, huh.NewOption(statusLabel(status), status))
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[apiclient.OrderStatus]().
				Title("Nouveau statut de "+order.Ref).
				Options(options...).
				Value(&fields.status),
		),
		huh.NewGroup(
			huh.NewInput().Title("Raison de l'annulation").Value(&fields.reason).Validate(required),
		).WithHideFunc(func() bool {
			return fields.status != apiclient.OrderCancelled
		}),
	).WithShowHelp(false)
	m.formKind = formStatus
	m.feedback = feedback{}
	return m, m.form.Init()
}

func (m ordersScreen) openOrderPickForm(users []user.User, products []product.Product) (ordersScreen, tea.Cmd) {
	if len(users) == 0 || len(products) == 0 {
		m.feedback = feedback{text: "il faut au moins un utilisateur et un produit", isError: true}
		return m, nil
	}
	m.customers = users
	m.catalog = products
	m.newOrder = &newOrderFields{}

	customerOptions := make([]huh.Option[int64], 0, len(users))
	for _, account := range users {
		customerOptions = append(customerOptions, huh.NewOption(account.Email, account.ID))
	}
	productOptions := make([]huh.Option[string], 0, len(products))
	for _, item := range products {
		productOptions = append(productOptions, huh.NewOption(item.Name+" ("+formatPrice(item.PriceTTCCents)+")", item.Ref))
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int64]().Title("Client").Options(customerOptions...).Value(&m.newOrder.customerID),
			huh.NewMultiSelect[string]().Title("Produits (espace pour cocher)").
				Options(productOptions...).
				Value(&m.newOrder.productRefs).
				Validate(func(refs []string) error {
					if len(refs) == 0 {
						return errors.New("choisis au moins un produit")
					}
					return nil
				}),
		),
	).WithShowHelp(false)
	m.formKind = formOrderPick
	m.feedback = feedback{}
	return m, m.form.Init()
}

func (m ordersScreen) openQuantitiesForm() (ordersScreen, tea.Cmd) {
	m.newOrder.quantities = make([]string, len(m.newOrder.productRefs))
	inputs := make([]huh.Field, 0, len(m.newOrder.productRefs))
	for index, productRef := range m.newOrder.productRefs {
		m.newOrder.quantities[index] = "1"
		inputs = append(inputs, huh.NewInput().
			Title("Quantité : "+m.productName(productRef)).
			Value(&m.newOrder.quantities[index]).
			Validate(validateQuantity))
	}

	m.form = huh.NewForm(huh.NewGroup(inputs...)).WithShowHelp(false)
	m.formKind = formOrderQuantities
	return m, m.form.Init()
}

func validateQuantity(value string) error {
	quantity, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || quantity < 1 || quantity > 1000 {
		return errors.New("entre 1 et 1000")
	}
	return nil
}

func (m ordersScreen) productName(productRef string) string {
	for _, item := range m.catalog {
		if item.Ref == productRef {
			return item.Name
		}
	}
	return productRef
}

func (m ordersScreen) updateForm(msg tea.Msg) (ordersScreen, tea.Cmd) {
	if isKey(msg, "esc") {
		m.form = nil
		m.formKind = formNone
		return m, nil
	}

	var cmd tea.Cmd
	m.form, cmd = updateForm(m.form, msg)
	if m.form.State != huh.StateCompleted {
		return m, cmd
	}

	kind := m.formKind
	m.form = nil
	m.formKind = formNone

	switch kind {
	case formStatus:
		fields := *m.status
		if fields.status != apiclient.OrderCancelled {
			fields.reason = ""
		}
		source := m.source
		return m, actionCmd(fields.orderRef+" → "+statusLabel(fields.status), func() error {
			return source.UpdateStatus(fields.orderRef, fields.status, fields.reason)
		})

	case formOrderPick:
		return m.openQuantitiesForm()

	case formOrderQuantities:
		draft := m.buildDraft()
		source := m.source
		return m, actionCmd("commande créée pour "+draft.customer.Email, func() error {
			return source.Create(draft)
		})
	}
	return m, nil
}

func (m ordersScreen) buildDraft() orderDraft {
	var draft orderDraft
	for _, account := range m.customers {
		if account.ID == m.newOrder.customerID {
			draft.customer = account
		}
	}
	for index, productRef := range m.newOrder.productRefs {
		quantity, _ := strconv.Atoi(strings.TrimSpace(m.newOrder.quantities[index])) // déjà validé
		for _, item := range m.catalog {
			if item.Ref == productRef {
				draft.lines = append(draft.lines, orderLine{product: item, quantity: quantity})
			}
		}
	}
	return draft
}

func (m ordersScreen) view() string {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Commandes") + "\n\n")
	if m.demo {
		builder.WriteString(warningStyle.Render("⚠ données fictives : routes /admin/orders pas encore disponibles côté API") + "\n\n")
	}

	if m.form != nil {
		builder.WriteString(m.form.View() + "\n\n")
		builder.WriteString(helpStyle.Render("entrée valider • échap annuler"))
		return builder.String()
	}
	if m.detail != nil {
		builder.WriteString(orderDetailView(*m.detail))
		return builder.String()
	}

	builder.WriteString(m.table.View() + "\n\n")
	builder.WriteString(m.feedback.view())
	builder.WriteString(helpStyle.Render(fmt.Sprintf("%d commande(s) • ↑/↓ naviguer • entrée détail • s statut • n nouvelle • échap menu", len(m.orders))))
	return builder.String()
}

func orderDetailView(order apiclient.Order) string {
	var builder strings.Builder
	builder.WriteString(selectedStyle.Render(order.Ref) + "\n\n")
	builder.WriteString("Client   " + order.UserEmail + "\n")
	builder.WriteString("Date     " + order.CreatedAt.Format("02/01/2006 15:04") + "\n")
	builder.WriteString("Statut   " + statusLabel(order.Status) + "\n")
	if order.CancelReason != "" {
		builder.WriteString("Raison   " + order.CancelReason + "\n")
	}
	builder.WriteString("\n")
	for _, item := range order.Items {
		fmt.Fprintf(&builder, "  %d × %-28s %s\n", item.Quantity, item.ProductName, formatPrice(item.UnitPriceTTCCents*int64(item.Quantity)))
	}
	builder.WriteString("\nTotal TTC " + formatPrice(order.TotalTTCCents) + "\n\n")
	builder.WriteString(helpStyle.Render("échap retour"))
	return builder.String()
}

func orderRows(orders []apiclient.Order) []table.Row {
	rows := make([]table.Row, 0, len(orders))
	for _, order := range orders {
		rows = append(rows, table.Row{
			order.Ref,
			order.UserEmail,
			statusLabel(order.Status),
			formatPrice(order.TotalTTCCents),
			order.CreatedAt.Format("02/01/2006 15:04"),
		})
	}
	return rows
}
