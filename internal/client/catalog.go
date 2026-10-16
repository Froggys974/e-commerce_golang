package client

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/product"
)

type searchResultMsg struct {
	query    string
	products []product.Product
	err      error
}

type catalog struct {
	api      *apiclient.Client
	search   textinput.Model
	table    table.Model
	products []product.Product
	selected *product.Product
	err      error
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

	return catalog{api: api, search: search, table: productTable}
}

func (m catalog) Init() tea.Cmd {
	return searchCmd(m.api, "")
}

func searchCmd(api *apiclient.Client, query string) tea.Cmd {
	return func() tea.Msg {
		products, err := api.SearchProducts(query)
		return searchResultMsg{query: query, products: products, err: err}
	}
}

func (m catalog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case searchResultMsg:
		// ignorer ancienne reponse qui arrive en retard
		if msg.query != m.search.Value() {
			return m, nil
		}
		m.err = msg.err
		m.products = msg.products
		m.table.SetRows(toRows(msg.products))
		m.table.SetCursor(0)
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.selected != nil {
			return m.updateDetail(msg)
		}
		return m.updateList(msg)
	}

	return m, nil
}

func (m catalog) updateList(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, tea.Quit
	case "enter":
		if len(m.products) > 0 {
			m.selected = &m.products[m.table.Cursor()]
		}
		return m, nil
	case "up", "down":
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	previousQuery := m.search.Value()
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	if m.search.Value() != previousQuery {
		return m, tea.Batch(cmd, searchCmd(m.api, m.search.Value()))
	}
	return m, cmd
}

func (m catalog) updateDetail(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.selected = nil
	}
	return m, nil
}

func (m catalog) View() tea.View {
	var screen string
	if m.selected != nil {
		screen = detailView(*m.selected)
	} else {
		screen = m.listView()
	}
	view := tea.NewView(screen)
	view.AltScreen = true
	return view
}

func (m catalog) listView() string {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Catalogue"))
	builder.WriteString("\n\n")
	builder.WriteString(m.search.View())
	builder.WriteString("\n\n")
	if m.err != nil {
		builder.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		builder.WriteString("\n\n")
	}
	builder.WriteString(m.table.View())
	builder.WriteString("\n\n")
	builder.WriteString(helpStyle.Render(fmt.Sprintf("%d produit(s) • ↑/↓ naviguer • entrée détail • échap quitter", len(m.products))))
	return builder.String()
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
	builder.WriteString(helpStyle.Render("échap retour"))
	return builder.String()
}

func toRows(products []product.Product) []table.Row {
	rows := make([]table.Row, 0, len(products))
	for _, item := range products {
		rows = append(rows, table.Row{item.Ref, item.Name, item.Category, formatPrice(item.PriceTTCCents)})
	}
	return rows
}
