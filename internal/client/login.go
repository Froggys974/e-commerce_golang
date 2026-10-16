package client

import (
	"errors"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
)

func askCredentials() (string, string, error) {
	var email, password string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Email").
				Value(&email).
				Validate(required),
			huh.NewInput().
				Title("Mot de passe").
				EchoMode(huh.EchoModePassword).
				Value(&password).
				Validate(required),
		),
	)

	err := form.Run()
	return email, password, err
}

func required(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("obligatoire")
	}
	return nil
}

func login(api *apiclient.Client) error {
	for {
		email, password, err := askCredentials()
		if err != nil {
			return err
		}

		err = api.Login(email, password)
		if err == nil {
			return nil
		}

		var apiError *apiclient.APIError
		if !errors.As(err, &apiError) {
			return err
		}
		lipgloss.Println(errorStyle.Render("✗ " + apiError.Message))
	}
}
