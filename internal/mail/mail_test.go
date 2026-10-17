package mail

import (
	"strings"
	"testing"
)

func TestBuildMessage(t *testing.T) {
	message, err := buildMessage("shop@test.local", "client@test.local", "Code de confirmation", "AB12CD34")
	if err != nil {
		t.Fatal(err)
	}
	text := string(message)
	for _, expected := range []string{"To: client@test.local\r\n", "\r\n\r\nBonjour", "AB12CD34"} {
		if !strings.Contains(text, expected) {
			t.Errorf("message sans %q :\n%s", expected, text)
		}
	}
}

func TestBuildMessageRejectsHeaderInjection(t *testing.T) {
	_, err := buildMessage("shop@test.local", "client@test.local\r\nBcc: victime@test.local", "Code", "X")
	if err == nil {
		t.Fatal("un \\r\\n dans l'adresse aurait dû être refusé")
	}
}
