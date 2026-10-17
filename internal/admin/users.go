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

func (m usersScreen) init() tea.Cmd {
	return loadUsersCmd(m.api)
}

func loadUsersCmd(api *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		users, err := api.ListUsers()
		return usersLoadedMsg{users: users, err: err}
	}
}

func (m usersScreen) update(msg tea.Msg) (usersScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case usersLoadedMsg:
		if msg.err != nil {
			m.feedback = feedback{text: msg.err.Error(), isError: true}
			return m, nil
		}
		m.users = msg.users
		m.table.SetRows(userRows(msg.users))
		if m.table.Cursor() >= len(msg.users) {
			m.table.SetCursor(0)
		}
		return m, nil

	case actionDoneMsg:
		m.feedback = feedbackFrom(msg)
		return m, loadUsersCmd(m.api)
	}

	if m.form != nil {
		return m.updateForm(msg)
	}
	if keyPress, ok := msg.(tea.KeyPressMsg); ok {
		return m.updateKeys(keyPress)
	}
	return m, nil
}

func (m usersScreen) updateKeys(keyPress tea.KeyPressMsg) (usersScreen, tea.Cmd) {
	key := keyPress.String()

	if m.confirmDelete {
		m.confirmDelete = false
		selected, ok := m.selected()
		if key != "o" || !ok {
			m.feedback = feedback{text: "suppression annulée"}
			return m, nil
		}
		return m, actionCmd("utilisateur "+selected.Email+" supprimé", func() error {
			return m.api.DeleteUser(selected.ID)
		})
	}

	switch key {
	case "esc":
		return m, openScreen(screenMenu)
	case "up", "down":
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(keyPress)
		return m, cmd
	case "n":
		return m.openNewUserForm()
	}

	selected, ok := m.selected()
	if !ok {
		return m, nil
	}
	isSelf := selected.ID == m.currentUserID

	switch key {
	case "c":
		if selected.Confirmed {
			m.feedback = feedback{text: "compte déjà confirmé"}
			return m, nil
		}
		return m, actionCmd("compte "+selected.Email+" confirmé", func() error {
			return m.api.ConfirmUser(selected.ID)
		})
	case "r":
		if isSelf {
			m.feedback = feedback{text: "impossible de changer ton propre rôle", isError: true}
			return m, nil
		}
		newRole := user.RoleAdmin
		if selected.Role == user.RoleAdmin {
			newRole = user.RoleUser
		}
		return m, actionCmd(selected.Email+" est maintenant "+string(newRole), func() error {
			return m.api.UpdateUser(selected.ID, selected.Email, newRole)
		})
	case "d":
		if isSelf {
			m.feedback = feedback{text: "impossible de supprimer ton propre compte", isError: true}
			return m, nil
		}
		m.confirmDelete = true
		m.feedback = feedback{text: "supprimer " + selected.Email + " ? o = oui, autre touche = non", isError: true}
		return m, nil
	}
	return m, nil
}

func (m usersScreen) selected() (user.User, bool) {
	if len(m.users) == 0 {
		return user.User{}, false
	}
	return m.users[m.table.Cursor()], true
}

func (m usersScreen) openNewUserForm() (usersScreen, tea.Cmd) {
	m.newUser = &newUserFields{role: user.RoleUser}
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Email").Value(&m.newUser.email).Validate(required),
			huh.NewInput().Title("Mot de passe (8 caractères min)").
				EchoMode(huh.EchoModePassword).
				Value(&m.newUser.password).
				Validate(minLength(8)),
			huh.NewSelect[user.Role]().Title("Rôle").
				Options(huh.NewOption("client", user.RoleUser), huh.NewOption("admin", user.RoleAdmin)).
				Value(&m.newUser.role),
		),
	).WithShowHelp(false)
	m.feedback = feedback{}
	return m, m.form.Init()
}

func minLength(size int) func(string) error {
	return func(value string) error {
		if len(value) < size {
			return fmt.Errorf("%d caractères minimum", size)
		}
		return nil
	}
}

func (m usersScreen) updateForm(msg tea.Msg) (usersScreen, tea.Cmd) {
	if isKey(msg, "esc") {
		m.form = nil
		return m, nil
	}

	var cmd tea.Cmd
	m.form, cmd = updateForm(m.form, msg)
	if m.form.State != huh.StateCompleted {
		return m, cmd
	}

	fields := *m.newUser
	m.form = nil
	return m, actionCmd("utilisateur "+fields.email+" créé", func() error {
		return m.api.CreateUser(fields.email, fields.password, fields.role)
	})
}

func (m usersScreen) view() string {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Utilisateurs") + "\n\n")

	if m.form != nil {
		builder.WriteString(m.form.View() + "\n\n")
		builder.WriteString(helpStyle.Render("entrée valider • échap annuler"))
		return builder.String()
	}

	builder.WriteString(m.table.View() + "\n\n")
	builder.WriteString(m.feedback.view())
	builder.WriteString(helpStyle.Render(fmt.Sprintf("%d utilisateur(s) • ↑/↓ naviguer • n nouveau • c confirmer • r changer rôle • d supprimer • échap menu", len(m.users))))
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
