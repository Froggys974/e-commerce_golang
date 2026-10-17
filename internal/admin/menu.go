package admin

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

type menuItem struct {
	label  string
	target screen
}

type menuScreen struct {
	items  []menuItem
	cursor int
}

func newMenuScreen() menuScreen {
	return menuScreen{
		items: []menuItem{
			{label: "Utilisateurs", target: screenUsers},
			{label: "Produits", target: screenProducts},
			{label: "Commandes", target: screenOrders},
		},
	}
}

func (m menuScreen) update(msg tea.Msg) (menuScreen, tea.Cmd) {
	keyPress, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch keyPress.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case "enter":
		return m, openScreen(m.items[m.cursor].target)
	case "q", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m menuScreen) view() string {
	var builder strings.Builder
	for index, item := range m.items {
		if index == m.cursor {
			builder.WriteString(selectedStyle.Render("> "+item.label) + "\n")
		} else {
			builder.WriteString("  " + item.label + "\n")
		}
	}
	builder.WriteString("\n" + helpStyle.Render("↑/↓ naviguer • entrée ouvrir • q quitter"))
	return builder.String()
}
