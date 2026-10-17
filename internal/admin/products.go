package admin

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/product"
)

type productsLoadedMsg struct {
	products []product.Product
	err      error
}

type newProductFields struct {
	name        string
	description string
	category    string
	price       string
	vatPercent  int
}

type productsScreen struct {
	api        *apiclient.Client
	table      table.Model
	products   []product.Product
	feedback   feedback
	form       *huh.Form
	newProduct *newProductFields
}

func newProductsScreen(api *apiclient.Client) productsScreen {
	productsTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "Réf", Width: 11},
			{Title: "Nom", Width: 26},
			{Title: "Catégorie", Width: 14},
			{Title: "Prix HT", Width: 12},
			{Title: "TVA", Width: 5},
			{Title: "Prix TTC", Width: 12},
		}),
		table.WithHeight(12),
		table.WithWidth(92),
		table.WithFocused(true),
	)
	return productsScreen{api: api, table: productsTable}
}

func (m productsScreen) init() tea.Cmd {
	return loadProductsCmd(m.api)
}

func loadProductsCmd(api *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		products, err := api.SearchProducts("")
		return productsLoadedMsg{products: products, err: err}
	}
}

func (m productsScreen) update(msg tea.Msg) (productsScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case productsLoadedMsg:
		if msg.err != nil {
			m.feedback = feedback{text: msg.err.Error(), isError: true}
			return m, nil
		}
		m.products = msg.products
		m.table.SetRows(productRows(msg.products))
		if m.table.Cursor() >= len(msg.products) {
			m.table.SetCursor(0)
		}
		return m, nil

	case actionDoneMsg:
		m.feedback = feedbackFrom(msg)
		return m, loadProductsCmd(m.api)
	}

	if m.form != nil {
		return m.updateForm(msg)
	}

	keyPress, ok := msg.(tea.KeyPressMsg)
	if !ok {
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
		return m.openNewProductForm()
	}
	return m, nil
}

func (m productsScreen) openNewProductForm() (productsScreen, tea.Cmd) {
	m.newProduct = &newProductFields{vatPercent: 20}
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Nom").Value(&m.newProduct.name).Validate(required),
			huh.NewInput().Title("Description").Value(&m.newProduct.description),
			huh.NewInput().Title("Catégorie").Value(&m.newProduct.category).Validate(required),
			huh.NewInput().Title("Prix HT en euros (ex : 19,99)").
				Value(&m.newProduct.price).
				Validate(func(value string) error {
					_, err := parseEuros(value)
					return err
				}),
			huh.NewSelect[int]().Title("TVA").
				Options(
					huh.NewOption("20 %", 20),
					huh.NewOption("10 %", 10),
					huh.NewOption("5 %", 5),
					huh.NewOption("0 %", 0),
				).
				Value(&m.newProduct.vatPercent),
		),
	).WithShowHelp(false)
	m.feedback = feedback{}
	return m, m.form.Init()
}

func (m productsScreen) updateForm(msg tea.Msg) (productsScreen, tea.Cmd) {
	if isKey(msg, "esc") {
		m.form = nil
		return m, nil
	}

	var cmd tea.Cmd
	m.form, cmd = updateForm(m.form, msg)
	if m.form.State != huh.StateCompleted {
		return m, cmd
	}

	fields := *m.newProduct
	m.form = nil
	priceCents, _ := parseEuros(fields.price)
	input := apiclient.CreateProductInput{
		Name:         fields.name,
		Description:  fields.description,
		Category:     fields.category,
		PriceHTCents: priceCents,
		VATPercent:   fields.vatPercent,
	}
	return m, actionCmd("produit "+fields.name+" créé", func() error {
		_, err := m.api.CreateProduct(input)
		return err
	})
}

func (m productsScreen) view() string {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Produits") + "\n\n")

	if m.form != nil {
		builder.WriteString(m.form.View() + "\n\n")
		builder.WriteString(helpStyle.Render("entrée valider • échap annuler"))
		return builder.String()
	}

	builder.WriteString(m.table.View() + "\n\n")
	if len(m.products) > 0 {
		builder.WriteString(helpStyle.Render("Description : "+m.products[m.table.Cursor()].Description) + "\n\n")
	}
	builder.WriteString(m.feedback.view())
	builder.WriteString(helpStyle.Render(fmt.Sprintf("%d produit(s) • ↑/↓ naviguer • n nouveau • échap menu", len(m.products))))
	return builder.String()
}

func productRows(products []product.Product) []table.Row {
	rows := make([]table.Row, 0, len(products))
	for _, item := range products {
		rows = append(rows, table.Row{
			item.Ref,
			item.Name,
			item.Category,
			formatPrice(item.PriceHTCents),
			fmt.Sprintf("%d %%", item.VATPercent),
			formatPrice(item.PriceTTCCents),
		})
	}
	return rows
}
