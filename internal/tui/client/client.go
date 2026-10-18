package client

import (
	"errors"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
)

func Run() error {
	lipgloss.Println(titleStyle.Render("E-commerce : connexion"))

	api := apiclient.New(apiclient.BaseURLFromEnv())
	for {
		action, err := askFirstAction()
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}

		switch action {
		case "login":
			err = login(api)
		case "register":
			err = register(api)
		case "forgot":
			err = forgotPassword(api)
		}
		if err != nil {
			return err
		}
		if action == "login" {
			break
		}
	}

	account, err := api.Me()
	if err != nil {
		return err
	}
	lipgloss.Println(successStyle.Render("✓ Connecté en tant que " + account.Email + " (" + string(account.Role) + ")"))
	_, err = tea.NewProgram(newCatalog(api)).Run()
	return err
}
