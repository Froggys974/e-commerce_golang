package client

import (
	tea "charm.land/bubbletea/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
)

type screen int

const (
	screenAuth screen = iota
	screenCatalog
)

type app struct {
	api     *apiclient.Client
	screen  screen
	auth    authScreen
	catalog catalog
}

func Run() error {
	_, err := tea.NewProgram(NewModel(apiclient.BaseURLFromEnv())).Run()
	return err
}

// un modele par session donc un client api et des tokens par utilisateur
func NewModel(apiURL string) tea.Model {
	return app{
		api:    apiclient.New(apiURL),
		screen: screenAuth,
		auth:   newAuthScreen(),
	}
}

func (model app) Init() tea.Cmd {
	return model.auth.form.Init()
}

func (model app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if isKey(msg, "ctrl+c") {
		return model, tea.Quit
	}
	if _, ok := msg.(loggedInMsg); ok {
		model.screen = screenCatalog
		model.catalog = newCatalog(model.api)
		return model, model.catalog.Init()
	}

	if model.screen == screenAuth {
		var cmd tea.Cmd
		model.auth, cmd = model.auth.update(msg, model.api)
		return model, cmd
	}

	updated, cmd := model.catalog.Update(msg)
	model.catalog = updated.(catalog)
	return model, cmd
}

func (model app) View() tea.View {
	if model.screen == screenCatalog {
		return model.catalog.View()
	}
	view := tea.NewView(titleStyle.Render("E-commerce") + "\n\n" + model.auth.view())
	view.AltScreen = true
	return view
}
