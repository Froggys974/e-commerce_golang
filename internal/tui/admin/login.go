package admin

import (
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

var errNotAdmin = errors.New("accès réservé aux administrateurs")

type loginResultMsg struct {
	account user.User
	err     error
}

type credentials struct {
	email    string
	password string
}

type loginScreen struct {
	form        *huh.Form
	credentials *credentials
	errMessage  string
	loading     bool
}

func newLoginScreen(errMessage string) loginScreen {
	creds := &credentials{}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Email").
				Value(&creds.email).
				Validate(required),
			huh.NewInput().
				Title("Mot de passe").
				EchoMode(huh.EchoModePassword).
				Value(&creds.password).
				Validate(required),
		),
	).WithShowHelp(false)

	return loginScreen{form: form, credentials: creds, errMessage: errMessage}
}

func (model loginScreen) update(msg tea.Msg, api *apiclient.Client) (loginScreen, tea.Cmd) {
	if model.loading {
		return model, nil
	}

	var cmd tea.Cmd
	model.form, cmd = updateForm(model.form, msg)

	if model.form.State == huh.StateCompleted {
		model.loading = true
		return model, tea.Batch(cmd, loginCmd(api, model.credentials.email, model.credentials.password))
	}
	return model, cmd
}

func loginCmd(api *apiclient.Client, email, password string) tea.Cmd {
	return func() tea.Msg {
		if err := api.Login(email, password); err != nil {
			return loginResultMsg{err: err}
		}
		account, err := api.Me()
		if err != nil {
			return loginResultMsg{err: err}
		}
		if account.Role != user.RoleAdmin {
			return loginResultMsg{err: errNotAdmin}
		}
		return loginResultMsg{account: account}
	}
}

func (model loginScreen) view() string {
	if model.loading {
		return "Connexion..."
	}
	var builder strings.Builder
	if model.errMessage != "" {
		builder.WriteString(errorStyle.Render("✗ "+model.errMessage) + "\n\n")
	}
	builder.WriteString(model.form.View() + "\n\n")
	builder.WriteString(helpStyle.Render("entrée valider • ctrl+c quitter"))
	return builder.String()
}
