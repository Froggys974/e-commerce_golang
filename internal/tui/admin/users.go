package admin

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

type usersLoadedMsg struct {
	users []user.User
	err   error
}

type newUserFields struct {
	email    string
	password string
	role     user.Role
}

type usersScreen struct {
	api           *apiclient.Client
	currentUserID int64
	table         table.Model
	users         []user.User
	feedback      feedback
	form          *huh.Form
	newUser       *newUserFields
	confirmDelete bool
}

func newUsersScreen(api *apiclient.Client, currentUserID int64) usersScreen {
	usersTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "ID", Width: 5},
			{Title: "Email", Width: 32},
			{Title: "Rôle", Width: 7},
			{Title: "Confirmé", Width: 9},
			{Title: "Créé le", Width: 11},
		}),
		table.WithHeight(12),
		table.WithWidth(76),
		table.WithFocused(true),
	)
	return usersScreen{api: api, currentUserID: currentUserID, table: usersTable}
}

func (model usersScreen) init() tea.Cmd {
	return loadUsersCmd(model.api)
}

func loadUsersCmd(api *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		users, err := api.ListUsers()
		return usersLoadedMsg{users: users, err: err}
	}
}

func (model usersScreen) update(msg tea.Msg) (usersScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case usersLoadedMsg:
		if msg.err != nil {
			model.feedback = feedback{text: msg.err.Error(), isError: true}
			return model, nil
		}
		model.users = msg.users
		model.table.SetRows(userRows(msg.users))
		if model.table.Cursor() >= len(msg.users) {
			model.table.SetCursor(0)
		}
		return model, nil

	case actionDoneMsg:
		model.feedback = feedbackFrom(msg)
		return model, loadUsersCmd(model.api)
	}

	if model.form != nil {
		return model.updateForm(msg)
	}
	if keyPress, ok := msg.(tea.KeyPressMsg); ok {
		return model.updateKeys(keyPress)
	}
	return model, nil
}

func (model usersScreen) updateKeys(keyPress tea.KeyPressMsg) (usersScreen, tea.Cmd) {
	key := keyPress.String()

	if model.confirmDelete {
		model.confirmDelete = false
		selected, ok := model.selected()
		if key != "o" || !ok {
			model.feedback = feedback{text: "suppression annulée"}
			return model, nil
		}
		return model, actionCmd("utilisateur "+selected.Email+" supprimé", func() error {
			return model.api.DeleteUser(selected.ID)
		})
	}

	switch key {
	case "esc":
		return model, openScreen(screenMenu)
	case "up", "down":
		var cmd tea.Cmd
		model.table, cmd = model.table.Update(keyPress)
		return model, cmd
	case "n":
		return model.openNewUserForm()
	}

	selected, ok := model.selected()
	if !ok {
		return model, nil
	}
	isSelf := selected.ID == model.currentUserID

	switch key {
	case "c":
		if selected.Confirmed {
			model.feedback = feedback{text: "compte déjà confirmé"}
			return model, nil
		}
		return model, actionCmd("compte "+selected.Email+" confirmé", func() error {
			return model.api.ConfirmUser(selected.ID)
		})
	case "r":
		if isSelf {
			model.feedback = feedback{text: "impossible de changer ton propre rôle", isError: true}
			return model, nil
		}
		newRole := user.RoleAdmin
		if selected.Role == user.RoleAdmin {
			newRole = user.RoleUser
		}
		return model, actionCmd(selected.Email+" est maintenant "+string(newRole), func() error {
			return model.api.UpdateUser(selected.ID, selected.Email, newRole)
		})
	case "d":
		if isSelf {
			model.feedback = feedback{text: "impossible de supprimer ton propre compte", isError: true}
			return model, nil
		}
		model.confirmDelete = true
		model.feedback = feedback{text: "supprimer " + selected.Email + " ? o = oui, autre touche = non", isError: true}
		return model, nil
	}
	return model, nil
}

func (model usersScreen) selected() (user.User, bool) {
	if len(model.users) == 0 {
		return user.User{}, false
	}
	return model.users[model.table.Cursor()], true
}

func (model usersScreen) openNewUserForm() (usersScreen, tea.Cmd) {
	model.newUser = &newUserFields{role: user.RoleUser}
	model.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Email").Value(&model.newUser.email).Validate(required),
			huh.NewInput().Title("Mot de passe (8 caractères min)").
				EchoMode(huh.EchoModePassword).
				Value(&model.newUser.password).
				Validate(minLength(8)),
			huh.NewSelect[user.Role]().Title("Rôle").
				Options(huh.NewOption("client", user.RoleUser), huh.NewOption("admin", user.RoleAdmin)).
				Value(&model.newUser.role),
		),
	).WithShowHelp(false)
	model.feedback = feedback{}
	return model, model.form.Init()
}

func minLength(size int) func(string) error {
	return func(value string) error {
		if len(value) < size {
			return fmt.Errorf("%d caractères minimum", size)
		}
		return nil
	}
}

func (model usersScreen) updateForm(msg tea.Msg) (usersScreen, tea.Cmd) {
	if isKey(msg, "esc") {
		model.form = nil
		return model, nil
	}

	var cmd tea.Cmd
	model.form, cmd = updateForm(model.form, msg)
	if model.form.State != huh.StateCompleted {
		return model, cmd
	}

	fields := *model.newUser
	model.form = nil
	return model, actionCmd("utilisateur "+fields.email+" créé", func() error {
		return model.api.CreateUser(fields.email, fields.password, fields.role)
	})
}

func (model usersScreen) view() string {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Utilisateurs") + "\n\n")

	if model.form != nil {
		builder.WriteString(model.form.View() + "\n\n")
		builder.WriteString(helpStyle.Render("entrée valider • échap annuler"))
		return builder.String()
	}

	builder.WriteString(model.table.View() + "\n\n")
	builder.WriteString(model.feedback.view())
	builder.WriteString(helpStyle.Render(fmt.Sprintf("%d utilisateur(s) • ↑/↓ naviguer • n nouveau • c confirmer • r changer rôle • d supprimer • échap menu", len(model.users))))
	return builder.String()
}

func userRows(users []user.User) []table.Row {
	rows := make([]table.Row, 0, len(users))
	for _, account := range users {
		confirmed := "non"
		if account.Confirmed {
			confirmed = "oui"
		}
		rows = append(rows, table.Row{
			strconv.FormatInt(account.ID, 10),
			account.Email,
			string(account.Role),
			confirmed,
			account.CreatedAt.Format("02/01/2006"),
		})
	}
	return rows
}
