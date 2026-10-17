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

func (m loginScreen) update(msg tea.Msg, api *apiclient.Client) (loginScreen, tea.Cmd) {
	if m.loading {
		return m, nil
	}

	var cmd tea.Cmd
	m.form, cmd = updateForm(m.form, msg)

	if m.form.State == huh.StateCompleted {
		m.loading = true
		return m, tea.Batch(cmd, loginCmd(api, m.credentials.email, m.credentials.password))
	}
	return m, cmd
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

func (m loginScreen) view() string {
	if m.loading {
		return "Connexion..."
	}
	var builder strings.Builder
	if m.errMessage != "" {
		builder.WriteString(errorStyle.Render("✗ "+m.errMessage) + "\n\n")
	}
	builder.WriteString(m.form.View() + "\n\n")
	builder.WriteString(helpStyle.Render("entrée valider • ctrl+c quitter"))
	return builder.String()
}
