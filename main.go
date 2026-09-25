package main

// commandes utiles :
//
// docker compose exec go go run .
// docker compose exec go go mod tidy
// go build -o go-test .

// pour la demo :
// https://www.terminal.shop/
// ssh terminal.shop

import (
	"fmt"
	"os"
	tea "charm.land/bubbletea/v2"
)

type model struct {
	cursor int
	counter int
}

func (m model) View() tea.View {
	content := fmt.Sprintf("Counter: %d\n\nAppuie sur [i] pour incrémenter, [d] pour décrémenter, [q] pour quitter.\n", m.counter)
	return tea.NewView(content)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
        switch msg.String() {
        case "i":
            m.counter += 1
        case "d":
            m.counter -= 1
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    }
	return m, nil
}

func initialModel() model {
	return model{
		counter: 0,
		cursor: 10,
	}
}

func main() {
    program := tea.NewProgram(initialModel())
    if _, err := program.Run(); err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
}