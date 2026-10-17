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

func (model menuScreen) update(msg tea.Msg) (menuScreen, tea.Cmd) {
	keyPress, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return model, nil
	}

	switch keyPress.String() {
	case "up", "k":
		if model.cursor > 0 {
			model.cursor--
		}
	case "down", "j":
		if model.cursor < len(model.items)-1 {
			model.cursor++
		}
	case "enter":
		return model, openScreen(model.items[model.cursor].target)
	case "q", "esc":
		return model, tea.Quit
	}
	return model, nil
}

func (model menuScreen) view() string {
	var builder strings.Builder
	for index, item := range model.items {
		if index == model.cursor {
			builder.WriteString(selectedStyle.Render("> "+item.label) + "\n")
		} else {
			builder.WriteString("  " + item.label + "\n")
		}
	}
	builder.WriteString("\n" + helpStyle.Render("↑/↓ naviguer • entrée ouvrir • q quitter"))
	return builder.String()
}
