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

func (m app) Init() tea.Cmd {
	return m.login.form.Init()
}

func (m app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if isKey(msg, "ctrl+c") {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case loginResultMsg:
		if msg.err != nil {
			m.login = newLoginScreen(msg.err.Error())
			return m, m.login.form.Init()
		}
		m.account = msg.account
		m.screen = screenMenu
		return m, nil

	case openScreenMsg:
		m.screen = msg.screen
		switch msg.screen {
		case screenUsers:
			m.users = newUsersScreen(m.api, m.account.ID)
			return m, m.users.init()
		case screenProducts:
			m.products = newProductsScreen(m.api)
			return m, m.products.init()
		case screenOrders:
			m.orders = newOrdersScreen(m.api)
			return m, m.orders.init()
		}
		return m, nil
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenLogin:
		m.login, cmd = m.login.update(msg, m.api)
	case screenMenu:
		m.menu, cmd = m.menu.update(msg)
	case screenUsers:
		m.users, cmd = m.users.update(msg)
	case screenProducts:
		m.products, cmd = m.products.update(msg)
	case screenOrders:
		m.orders, cmd = m.orders.update(msg)
	}
	return m, cmd
}

func (m app) View() tea.View {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("E-commerce : administration"))
	if m.screen != screenLogin {
		builder.WriteString(helpStyle.Render("  connecté : " + m.account.Email))
	}
	builder.WriteString("\n\n")

	switch m.screen {
	case screenLogin:
		builder.WriteString(m.login.view())
	case screenMenu:
		builder.WriteString(m.menu.view())
	case screenUsers:
		builder.WriteString(m.users.view())
	case screenProducts:
		builder.WriteString(m.products.view())
	case screenOrders:
		builder.WriteString(m.orders.view())
	}

	view := tea.NewView(builder.String())
	view.AltScreen = true
	return view
}
