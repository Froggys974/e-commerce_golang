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

func (model productsScreen) init() tea.Cmd {
	return loadProductsCmd(model.api)
}

func loadProductsCmd(api *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		products, err := api.SearchProducts("")
		return productsLoadedMsg{products: products, err: err}
	}
}

func (model productsScreen) update(msg tea.Msg) (productsScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case productsLoadedMsg:
		if msg.err != nil {
			model.feedback = feedback{text: msg.err.Error(), isError: true}
			return model, nil
		}
		model.products = msg.products
		model.table.SetRows(productRows(msg.products))
		if model.table.Cursor() >= len(msg.products) {
			model.table.SetCursor(0)
		}
		return model, nil

	case actionDoneMsg:
		model.feedback = feedbackFrom(msg)
		return model, loadProductsCmd(model.api)
	}

	if model.form != nil {
		return model.updateForm(msg)
	}

	keyPress, ok := msg.(tea.KeyPressMsg)
	if !ok {
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
		return model.openNewProductForm()
	}
	return model, nil
}

func (model productsScreen) openNewProductForm() (productsScreen, tea.Cmd) {
	model.newProduct = &newProductFields{vatPercent: 20}
	model.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Nom").Value(&model.newProduct.name).Validate(required),
			huh.NewInput().Title("Description").Value(&model.newProduct.description),
			huh.NewInput().Title("Catégorie").Value(&model.newProduct.category).Validate(required),
			huh.NewInput().Title("Prix HT en euros (ex : 19,99)").
				Value(&model.newProduct.price).
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
				Value(&model.newProduct.vatPercent),
		),
	).WithShowHelp(false)
	model.feedback = feedback{}
	return model, model.form.Init()
}

func (model productsScreen) updateForm(msg tea.Msg) (productsScreen, tea.Cmd) {
	if isKey(msg, "esc") {
		model.form = nil
		return model, nil
	}

	var cmd tea.Cmd
	model.form, cmd = updateForm(model.form, msg)
	if model.form.State != huh.StateCompleted {
		return model, cmd
	}

	fields := *model.newProduct
	model.form = nil
	priceCents, _ := parseEuros(fields.price)
	input := apiclient.CreateProductInput{
		Name:         fields.name,
		Description:  fields.description,
		Category:     fields.category,
		PriceHTCents: priceCents,
		VATPercent:   fields.vatPercent,
	}
	return model, actionCmd("produit "+fields.name+" créé", func() error {
		_, err := model.api.CreateProduct(input)
		return err
	})
}

func (model productsScreen) view() string {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Produits") + "\n\n")

	if model.form != nil {
		builder.WriteString(model.form.View() + "\n\n")
		builder.WriteString(helpStyle.Render("entrée valider • échap annuler"))
		return builder.String()
	}

	builder.WriteString(model.table.View() + "\n\n")
	if len(model.products) > 0 {
		builder.WriteString(helpStyle.Render("Description : "+model.products[model.table.Cursor()].Description) + "\n\n")
	}
	builder.WriteString(model.feedback.view())
	builder.WriteString(helpStyle.Render(fmt.Sprintf("%d produit(s) • ↑/↓ naviguer • n nouveau • échap menu", len(model.products))))
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
