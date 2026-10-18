package client

import (
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

type feedback struct {
	text    string
	isError bool
}

func (message feedback) view() string {
	switch {
	case message.text == "":
		return ""
	case message.isError:
		return errorStyle.Render("✗ "+message.text) + "\n\n"
	default:
		return successStyle.Render("✓ "+message.text) + "\n\n"
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
