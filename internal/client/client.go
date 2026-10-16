package client

import (
	"errors"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
)

const defaultAPIURL = "http://localhost:8000"

func Run() error {
	lipgloss.Println(titleStyle.Render("E-commerce : connexion"))

	api := apiclient.New(apiURL())
	if err := login(api); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	account, err := api.Me()
	if err != nil {
		return err
	}
	lipgloss.Println(successStyle.Render("✓ Connecté en tant que " + account.Email + " (" + string(account.Role) + ")"))
	_, err = tea.NewProgram(newCatalog(api)).Run()
	return err
}

// API_URL ==> pointer vers ngrok url pour la soutenance
func apiURL() string {
	if url := os.Getenv("API_URL"); url != "" {
		return url
	}
	return defaultAPIURL
}
