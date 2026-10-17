package admin

import (
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

type actionDoneMsg struct {
	message string
	err     error
}

func actionCmd(successMessage string, action func() error) tea.Cmd {
	return func() tea.Msg {
		if err := action(); err != nil {
			return actionDoneMsg{err: err}
		}
		return actionDoneMsg{message: successMessage}
	}
}

type feedback struct {
	text    string
	isError bool
}

func feedbackFrom(msg actionDoneMsg) feedback {
	if msg.err != nil {
		return feedback{text: msg.err.Error(), isError: true}
	}
	return feedback{text: msg.message}
}

func (f feedback) view() string {
	switch {
	case f.text == "":
		return ""
	case f.isError:
		return errorStyle.Render("✗ "+f.text) + "\n\n"
	default:
		return successStyle.Render("✓ "+f.text) + "\n\n"
	}
}

func updateForm(form *huh.Form, msg tea.Msg) (*huh.Form, tea.Cmd) {
	model, cmd := form.Update(msg)
	if updated, ok := model.(*huh.Form); ok {
		form = updated
	}
	return form, cmd
}

func required(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("obligatoire")
	}
	return nil
}

func isKey(msg tea.Msg, key string) bool {
	keyPress, ok := msg.(tea.KeyPressMsg)
	return ok && keyPress.String() == key
}
