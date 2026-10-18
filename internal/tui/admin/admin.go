package admin

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type screen int

const (
	screenLogin screen = iota
	screenMenu
	screenUsers
	screenProducts
	screenOrders
)

// envoyé par un écran pour demander à l'app d'en afficher un autre
type openScreenMsg struct {
	screen screen
}

func openScreen(target screen) tea.Cmd {
	return func() tea.Msg {
		return openScreenMsg{screen: target}
	}
}

type app struct {
	api      *apiclient.Client
	screen   screen
	account  user.User
	login    loginScreen
	menu     menuScreen
	users    usersScreen
	products productsScreen
	orders   ordersScreen
}

func Run() error {
	_, err := tea.NewProgram(NewModel(apiclient.BaseURLFromEnv())).Run()
	return err
}

func NewModel(apiURL string) tea.Model {
	return app{
		api:    apiclient.New(apiURL),
		screen: screenLogin,
		login:  newLoginScreen(""),
		menu:   newMenuScreen(),
	}
}

func (model app) Init() tea.Cmd {
	return model.login.form.Init()
}

func (model app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if isKey(msg, "ctrl+c") {
		return model, tea.Quit
	}

	switch msg := msg.(type) {
	case loginResultMsg:
		if msg.err != nil {
			model.login = newLoginScreen(msg.err.Error())
			return model, model.login.form.Init()
		}
		model.account = msg.account
		model.screen = screenMenu
		return model, nil

	case openScreenMsg:
		model.screen = msg.screen
		switch msg.screen {
		case screenUsers:
			model.users = newUsersScreen(model.api, model.account.ID)
			return model, model.users.init()
		case screenProducts:
			model.products = newProductsScreen(model.api)
			return model, model.products.init()
		case screenOrders:
			model.orders = newOrdersScreen(model.api)
			return model, model.orders.init()
		}
		return model, nil
	}

	var cmd tea.Cmd
	switch model.screen {
	case screenLogin:
		model.login, cmd = model.login.update(msg, model.api)
	case screenMenu:
		model.menu, cmd = model.menu.update(msg)
	case screenUsers:
		model.users, cmd = model.users.update(msg)
	case screenProducts:
		model.products, cmd = model.products.update(msg)
	case screenOrders:
		model.orders, cmd = model.orders.update(msg)
	}
	return model, cmd
}

func (model app) View() tea.View {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("E-commerce : administration"))
	if model.screen != screenLogin {
		builder.WriteString(helpStyle.Render("  connecté : " + model.account.Email))
	}
	builder.WriteString("\n\n")

	switch model.screen {
	case screenLogin:
		builder.WriteString(model.login.view())
	case screenMenu:
		builder.WriteString(model.menu.view())
	case screenUsers:
		builder.WriteString(model.users.view())
	case screenProducts:
		builder.WriteString(model.products.view())
	case screenOrders:
		builder.WriteString(model.orders.view())
	}

	view := tea.NewView(builder.String())
	view.AltScreen = true
	return view
}
