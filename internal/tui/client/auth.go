package client

import (
	"errors"
	"net/http"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
)

func askFirstAction() (string, error) {
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Bienvenue").
				Options(
					huh.NewOption("Se connecter", "login"),
					huh.NewOption("Créer un compte", "register"),
					huh.NewOption("Mot de passe oublié", "forgot"),
				).
				Value(&action),
		),
	)

	err := form.Run()
	return action, err
}

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
		if !errors.As(err, &apiError) || apiError.Status != http.StatusUnauthorized {
			return err
		}
		lipgloss.Println(errorStyle.Render("✗ " + apiError.Message))
	}
}

func register(api *apiclient.Client) error {
	email, password, err := askCredentials()
	if err != nil {
		return err
	}
	if err := api.Register(email, password); err != nil {
		return showAPIError(err)
	}
	lipgloss.Println(successStyle.Render("✓ Un code de confirmation a été envoyé à " + email))

	for {
		code, err := askCode(email)
		if err != nil {
			return err
		}

		err = api.Confirm(email, code)
		if err == nil {
			lipgloss.Println(successStyle.Render("✓ Compte confirmé, tu peux te connecter"))
			return nil
		}
		if err := showAPIError(err); err != nil {
			return err
		}

		resend, err := askYesNo("Renvoyer un nouveau code ?")
		if err != nil {
			return err
		}
		if resend {
			if err := api.ResendConfirmCode(email); err != nil {
				return showAPIError(err)
			}
			lipgloss.Println(successStyle.Render("✓ Nouveau code envoyé à " + email))
		}
	}
}

func forgotPassword(api *apiclient.Client) error {
	var email string
	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Email du compte").Value(&email).Validate(required),
	)).Run()
	if err != nil {
		return err
	}
	if err := api.ForgotPassword(email); err != nil {
		return showAPIError(err)
	}
	lipgloss.Println(successStyle.Render("✓ Si un compte existe, un code a été envoyé à " + email))

	for {
		code, newPassword, err := askResetForm(email)
		if err != nil {
			return err
		}

		err = api.ResetPassword(email, code, newPassword)
		if err == nil {
			lipgloss.Println(successStyle.Render("✓ Mot de passe modifié, tu peux te connecter"))
			return nil
		}
		if err := showAPIError(err); err != nil {
			return err
		}

		resend, err := askYesNo("Renvoyer un nouveau code ?")
		if err != nil {
			return err
		}
		if resend {
			if err := api.ForgotPassword(email); err != nil {
				return showAPIError(err)
			}
			lipgloss.Println(successStyle.Render("✓ Nouveau code envoyé à " + email))
		}
	}
}

func askCode(email string) (string, error) {
	var code string
	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Code reçu").
			Description("Saisis le code envoyé à " + email + " (valable 15 minutes)").
			Value(&code).
			Validate(required),
	)).Run()
	return code, err
}

func askResetForm(email string) (string, string, error) {
	var code, newPassword string
	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Code reçu").
			Description("Saisis le code envoyé à "+email+" (valable 15 minutes)").
			Value(&code).
			Validate(required),
		huh.NewInput().
			Title("Nouveau mot de passe").
			Description("8 caractères minimum").
			EchoMode(huh.EchoModePassword).
			Value(&newPassword).
			Validate(required),
	)).Run()
	return code, newPassword, err
}

func askYesNo(question string) (bool, error) {
	var answer bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(question).
			Affirmative("Oui").
			Negative("Non, réessayer").
			Value(&answer),
	)).Run()
	return answer, err
}

// erreur api affichee et on continue sinon erreur remontee
func showAPIError(err error) error {
	var apiError *apiclient.APIError
	if !errors.As(err, &apiError) {
		return err
	}
	lipgloss.Println(errorStyle.Render("✗ " + apiError.Message))
	return nil
}
