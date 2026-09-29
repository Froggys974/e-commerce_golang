package client

import (
    "fmt"
    tea "charm.land/bubbletea/v2"
)

type model struct {
    counter int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        switch msg.String() {
        case "i":
            m.counter++
        case "d":
            m.counter--
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m model) View() tea.View {
    return tea.NewView(fmt.Sprintf("Counter: %d\n\n[i] +1  [d] -1  [q] quitter\n", m.counter))
}

func Run() error {
    _, err := tea.NewProgram(model{}).Run()
    return err
}