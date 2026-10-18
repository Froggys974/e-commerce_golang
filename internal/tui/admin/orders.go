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
	"github.com/Froggys974/e-commerce_golang/internal/order"
	"github.com/Froggys974/e-commerce_golang/internal/product"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type ordersLoadedMsg struct {
	orders []order.Order
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
	status   order.Status
	reason   string
}

type newOrderFields struct {
	customerID  int64
	productRefs []string
	quantities  []string
}

type orderDraft struct {
	customer user.User
	lines    []orderLine
}

type orderLine struct {
	product  product.Product
	quantity int
}

func (draft orderDraft) items() []order.ItemInput {
	items := make([]order.ItemInput, 0, len(draft.lines))
	for _, line := range draft.lines {
		items = append(items, order.ItemInput{ProductRef: line.product.Ref, Quantity: line.quantity})
	}
	return items
}

type ordersScreen struct {
	api      *apiclient.Client
	table    table.Model
	orders   []order.Order
	detail   *order.Order
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
	return ordersScreen{api: api, table: ordersTable}
}

func (model ordersScreen) init() tea.Cmd {
	return loadOrdersCmd(model.api)
}

func loadOrdersCmd(api *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		orders, err := api.ListAllOrders()
		return ordersLoadedMsg{orders: orders, err: err}
	}
}

func loadOrderFormDataCmd(api *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		users, err := api.ListUsers()
		if err != nil {
			return orderFormDataMsg{err: err}
		}
		products, err := api.SearchProducts(product.SearchFilters{})
		return orderFormDataMsg{users: users, products: products, err: err}
	}
}

func (model ordersScreen) update(msg tea.Msg) (ordersScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case ordersLoadedMsg:
		if msg.err != nil {
			model.feedback = feedback{text: msg.err.Error(), isError: true}
			return model, nil
		}
		model.orders = msg.orders
		model.table.SetRows(orderRows(msg.orders))
		if model.table.Cursor() >= len(msg.orders) {
			model.table.SetCursor(0)
		}
		return model, nil

	case orderFormDataMsg:
		if msg.err != nil {
			model.feedback = feedback{text: msg.err.Error(), isError: true}
			return model, nil
		}
		return model.openOrderPickForm(msg.users, msg.products)

	case actionDoneMsg:
		model.feedback = feedbackFrom(msg)
		return model, loadOrdersCmd(model.api)
	}

	if model.form != nil {
		return model.updateForm(msg)
	}

	keyPress, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return model, nil
	}
	if model.detail != nil {
		if keyPress.String() == "esc" {
			model.detail = nil
		}
		return model, nil
	}

	switch keyPress.String() {
	case "esc":
		return model, openScreen(screenMenu)
	case "up", "down":
		var cmd tea.Cmd
		model.table, cmd = model.table.Update(keyPress)
		return model, cmd
	case "n":
		model.feedback = feedback{text: "chargement des clients et produits..."}
		return model, loadOrderFormDataCmd(model.api)
	}

	if len(model.orders) == 0 {
		return model, nil
	}
	selected := model.orders[model.table.Cursor()]
	switch keyPress.String() {
	case "enter":
		model.detail = &selected
	case "s":
		return model.openStatusForm(selected)
	}
	return model, nil
}

func (model ordersScreen) openStatusForm(current order.Order) (ordersScreen, tea.Cmd) {
	options := make([]huh.Option[order.Status], 0, len(order.Statuses))
	for _, status := range order.Statuses {
		if current.Status.CanBecome(status) {
			options = append(options, huh.NewOption(statusLabel(status), status))
		}
	}
	if len(options) == 0 {
		model.feedback = feedback{text: current.Ref + " est " + statusLabel(current.Status) + " : statut définitif", isError: true}
		return model, nil
	}

	fields := &statusFields{orderRef: current.Ref, status: options[0].Value}
	model.status = fields

	model.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[order.Status]().
				Title("Nouveau statut de "+current.Ref).
				Options(options...).
				Value(&fields.status),
		),
		huh.NewGroup(
			huh.NewInput().Title("Raison de l'annulation").Value(&fields.reason).Validate(required),
		).WithHideFunc(func() bool {
			return fields.status != order.StatusCancelled
		}),
	).WithShowHelp(false)
	model.formKind = formStatus
	model.feedback = feedback{}
	return model, model.form.Init()
}

func (model ordersScreen) openOrderPickForm(users []user.User, products []product.Product) (ordersScreen, tea.Cmd) {
	if len(users) == 0 || len(products) == 0 {
		model.feedback = feedback{text: "il faut au moins un utilisateur et un produit", isError: true}
		return model, nil
	}
	model.customers = users
	model.catalog = products
	model.newOrder = &newOrderFields{}

	customerOptions := make([]huh.Option[int64], 0, len(users))
	for _, account := range users {
		customerOptions = append(customerOptions, huh.NewOption(account.Email, account.ID))
	}
	productOptions := make([]huh.Option[string], 0, len(products))
	for _, item := range products {
		productOptions = append(productOptions, huh.NewOption(item.Name+" ("+formatPrice(item.PriceTTCCents)+")", item.Ref))
	}

	model.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int64]().Title("Client").Options(customerOptions...).Value(&model.newOrder.customerID),
			huh.NewMultiSelect[string]().Title("Produits (espace pour cocher)").
				Options(productOptions...).
				Value(&model.newOrder.productRefs).
				Validate(func(refs []string) error {
					if len(refs) == 0 {
						return errors.New("choisis au moins un produit")
					}
					return nil
				}),
		),
	).WithShowHelp(false)
	model.formKind = formOrderPick
	model.feedback = feedback{}
	return model, model.form.Init()
}

func (model ordersScreen) openQuantitiesForm() (ordersScreen, tea.Cmd) {
	model.newOrder.quantities = make([]string, len(model.newOrder.productRefs))
	inputs := make([]huh.Field, 0, len(model.newOrder.productRefs))
	for index, productRef := range model.newOrder.productRefs {
		model.newOrder.quantities[index] = "1"
		inputs = append(inputs, huh.NewInput().
			Title("Quantité : "+model.productName(productRef)).
			Value(&model.newOrder.quantities[index]).
			Validate(validateQuantity))
	}

	model.form = huh.NewForm(huh.NewGroup(inputs...)).WithShowHelp(false)
	model.formKind = formOrderQuantities
	return model, model.form.Init()
}

// meme limite que l api
func validateQuantity(value string) error {
	quantity, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || quantity < 1 || quantity > 99 {
		return errors.New("entre 1 et 99")
	}
	return nil
}

func (model ordersScreen) productName(productRef string) string {
	for _, item := range model.catalog {
		if item.Ref == productRef {
			return item.Name
		}
	}
	return productRef
}

func (model ordersScreen) updateForm(msg tea.Msg) (ordersScreen, tea.Cmd) {
	if isKey(msg, "esc") {
		model.form = nil
		model.formKind = formNone
		return model, nil
	}

	var cmd tea.Cmd
	model.form, cmd = updateForm(model.form, msg)
	if model.form.State != huh.StateCompleted {
		return model, cmd
	}

	kind := model.formKind
	model.form = nil
	model.formKind = formNone

	switch kind {
	case formStatus:
		fields := *model.status
		if fields.status != order.StatusCancelled {
			fields.reason = ""
		}
		api := model.api
		return model, actionCmd(fields.orderRef+" → "+statusLabel(fields.status), func() error {
			return api.UpdateOrderStatus(fields.orderRef, fields.status, fields.reason)
		})

	case formOrderPick:
		return model.openQuantitiesForm()

	case formOrderQuantities:
		draft := model.buildDraft()
		api := model.api
		return model, actionCmd("commande créée pour "+draft.customer.Email, func() error {
			return api.CreateOrder(draft.customer.ID, draft.items())
		})
	}
	return model, nil
}

func (model ordersScreen) buildDraft() orderDraft {
	var draft orderDraft
	for _, account := range model.customers {
		if account.ID == model.newOrder.customerID {
			draft.customer = account
		}
	}
	for index, productRef := range model.newOrder.productRefs {
		// deja verifie par validateQuantity
		quantity, _ := strconv.Atoi(strings.TrimSpace(model.newOrder.quantities[index]))
		for _, item := range model.catalog {
			if item.Ref == productRef {
				draft.lines = append(draft.lines, orderLine{product: item, quantity: quantity})
			}
		}
	}
	return draft
}

func (model ordersScreen) view() string {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Commandes") + "\n\n")

	if model.form != nil {
		builder.WriteString(model.form.View() + "\n\n")
		builder.WriteString(helpStyle.Render("entrée valider • échap annuler"))
		return builder.String()
	}
	if model.detail != nil {
		builder.WriteString(orderDetailView(*model.detail))
		return builder.String()
	}

	builder.WriteString(model.table.View() + "\n\n")
	builder.WriteString(model.feedback.view())
	builder.WriteString(helpStyle.Render(fmt.Sprintf("%d commande(s) • ↑/↓ naviguer • entrée détail • s statut • n nouvelle • échap menu", len(model.orders))))
	return builder.String()
}

func orderDetailView(current order.Order) string {
	var builder strings.Builder
	builder.WriteString(selectedStyle.Render(current.Ref) + "\n\n")
	builder.WriteString("Client   " + current.UserEmail + "\n")
	builder.WriteString("Date     " + current.CreatedAt.Local().Format("02/01/2006 15:04") + "\n")
	builder.WriteString("Statut   " + statusLabel(current.Status) + "\n")
	if current.CancelReason != "" {
		builder.WriteString("Raison   " + current.CancelReason + "\n")
	}
	builder.WriteString("\n")
	for _, item := range current.Items {
		fmt.Fprintf(&builder, "  %d × %-28s %s\n", item.Quantity, item.ProductName, formatPrice(item.UnitPriceTTCCents*int64(item.Quantity)))
	}
	builder.WriteString("\nTotal TTC " + formatPrice(current.TotalTTCCents) + "\n\n")
	builder.WriteString(helpStyle.Render("échap retour"))
	return builder.String()
}

func orderRows(orders []order.Order) []table.Row {
	rows := make([]table.Row, 0, len(orders))
	for _, current := range orders {
		rows = append(rows, table.Row{
			current.Ref,
			current.UserEmail,
			statusLabel(current.Status),
			formatPrice(current.TotalTTCCents),
			current.CreatedAt.Local().Format("02/01/2006 15:04"),
		})
	}
	return rows
}
