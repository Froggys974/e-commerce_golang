package client

import (
	"errors"
	"net/http"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
)

type authStep int

const (
	stepChoice authStep = iota
	stepLogin
	stepRegister
	stepConfirm
	stepForgot
	stepReset
)

type authResultMsg struct {
	next    authStep
	message string
	isError bool
	err     error
}

type loggedInMsg struct{}

// pointeur partage avec les formulaires huh qui ecrivent dedans
type authFields struct {
	action   string
	email    string
	password string
	code     string
}

type authScreen struct {
	step     authStep
	form     *huh.Form
	fields   *authFields
	feedback feedback
	loading  bool
}

func newAuthScreen() authScreen {
	fields := &authFields{}
	return authScreen{step: stepChoice, form: formFor(stepChoice, fields), fields: fields}
}

func (screen authScreen) goTo(step authStep, message feedback) (authScreen, tea.Cmd) {
	screen.fields.password = ""
	screen.fields.code = ""
	screen.step = step
	screen.feedback = message
	screen.loading = false
	screen.form = formFor(step, screen.fields)
	return screen, screen.form.Init()
}

func (screen authScreen) update(msg tea.Msg, api *apiclient.Client) (authScreen, tea.Cmd) {
	if result, ok := msg.(authResultMsg); ok {
		if result.err != nil {
			return screen.goTo(screen.step, feedback{text: result.err.Error(), isError: true})
		}
		return screen.goTo(result.next, feedback{text: result.message, isError: result.isError})
	}
	if screen.loading {
		return screen, nil
	}

	if isKey(msg, "esc") {
		if screen.step == stepChoice {
			return screen, tea.Quit
		}
		return screen.goTo(stepChoice, feedback{})
	}
	if isKey(msg, "ctrl+r") && (screen.step == stepConfirm || screen.step == stepReset) {
		screen.loading = true
		return screen, resendCmd(api, screen.step, screen.fields.email)
	}

	var cmd tea.Cmd
	screen.form, cmd = updateForm(screen.form, msg)
	if screen.form.State != huh.StateCompleted {
		return screen, cmd
	}

	fields := screen.fields
	switch screen.step {
	case stepChoice:
		return screen.goTo(choiceToStep(fields.action), feedback{})
	case stepLogin:
		screen.loading = true
		return screen, loginCmd(api, fields.email, fields.password)
	case stepRegister:
		screen.loading = true
		return screen, authCmd(stepConfirm, "Un code de confirmation a été envoyé à "+fields.email, func() error {
			return api.Register(fields.email, fields.password)
		})
	case stepConfirm:
		screen.loading = true
		return screen, authCmd(stepLogin, "Compte confirmé, tu peux te connecter", func() error {
			return api.Confirm(fields.email, fields.code)
		})
	case stepForgot:
		screen.loading = true
		return screen, authCmd(stepReset, "Si un compte existe, un code a été envoyé à "+fields.email, func() error {
			return api.ForgotPassword(fields.email)
		})
	case stepReset:
		screen.loading = true
		return screen, authCmd(stepLogin, "Mot de passe modifié, tu peux te connecter", func() error {
			return api.ResetPassword(fields.email, fields.code, fields.password)
		})
	}
	return screen, cmd
}

func choiceToStep(action string) authStep {
	switch action {
	case "register":
		return stepRegister
	case "forgot":
		return stepForgot
	}
	return stepLogin
}

func authCmd(next authStep, message string, action func() error) tea.Cmd {
	return func() tea.Msg {
		if err := action(); err != nil {
			return authResultMsg{err: err}
		}
		return authResultMsg{next: next, message: message}
	}
}

func loginCmd(api *apiclient.Client, email, password string) tea.Cmd {
	return func() tea.Msg {
		err := api.Login(email, password)
		var apiError *apiclient.APIError
		if errors.As(err, &apiError) && apiError.Status == http.StatusForbidden {
			return authResultMsg{next: stepConfirm, message: "Compte non confirmé : saisis le code reçu ou ctrl+r pour en recevoir un nouveau", isError: true}
		}
		if err != nil {
			return authResultMsg{err: err}
		}
		return loggedInMsg{}
	}
}

func resendCmd(api *apiclient.Client, step authStep, email string) tea.Cmd {
	return authCmd(step, "Nouveau code envoyé à "+email, func() error {
		if step == stepReset {
			return api.ForgotPassword(email)
		}
		return api.ResendConfirmCode(email)
	})
}

func formFor(step authStep, fields *authFields) *huh.Form {
	var group *huh.Group
	switch step {
	case stepChoice:
		group = huh.NewGroup(
			huh.NewSelect[string]().
				Title("Bienvenue").
				Options(
					huh.NewOption("Se connecter", "login"),
					huh.NewOption("Créer un compte", "register"),
					huh.NewOption("Mot de passe oublié", "forgot"),
				).
				Value(&fields.action),
		)
	case stepLogin, stepRegister:
		title := "Connexion"
		if step == stepRegister {
			title = "Inscription"
		}
		group = huh.NewGroup(
			huh.NewInput().Title("Email").Value(&fields.email).Validate(required),
			huh.NewInput().
				Title("Mot de passe").
				Description("8 caractères minimum").
				EchoMode(huh.EchoModePassword).
				Value(&fields.password).
				Validate(required),
		).Title(title)
	case stepConfirm:
		group = huh.NewGroup(
			codeInput(fields),
		).Title("Confirmation du compte")
	case stepForgot:
		group = huh.NewGroup(
			huh.NewInput().Title("Email du compte").Value(&fields.email).Validate(required),
		).Title("Mot de passe oublié")
	case stepReset:
		group = huh.NewGroup(
			codeInput(fields),
			huh.NewInput().
				Title("Nouveau mot de passe").
				Description("8 caractères minimum").
				EchoMode(huh.EchoModePassword).
				Value(&fields.password).
				Validate(required),
		).Title("Nouveau mot de passe")
	}
	return huh.NewForm(group).WithShowHelp(false)
}

func codeInput(fields *authFields) *huh.Input {
	return huh.NewInput().
		Title("Code reçu").
		Description("Envoyé à " + fields.email + " (valable 15 minutes)").
		Value(&fields.code).
		Validate(required)
}

func (screen authScreen) view() string {
	if screen.loading {
		return "Chargement..."
	}

	help := "entrée valider • échap retour • ctrl+c quitter"
	switch screen.step {
	case stepChoice:
		help = "↑/↓ choisir • entrée valider • échap quitter"
	case stepConfirm, stepReset:
		help = "entrée valider • ctrl+r renvoyer un code • échap retour"
	}

	var builder strings.Builder
	builder.WriteString(screen.feedback.view())
	builder.WriteString(screen.form.View() + "\n\n")
	builder.WriteString(helpStyle.Render(help))
	return builder.String()
}
