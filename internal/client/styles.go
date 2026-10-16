package client

import "charm.land/lipgloss/v2"

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#9300c5"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)
