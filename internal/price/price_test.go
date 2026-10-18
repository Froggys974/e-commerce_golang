package price

import "testing"

func TestParseEuros(t *testing.T) {
	valid := map[string]int64{
		"19,99":  1999,
		"19.99":  1999,
		"19":     1900,
		"19,9":   1990,
		" 0,05 ": 5,
		"0":      0,
	}
	for input, expected := range valid {
		got, err := ParseEuros(input)
		if err != nil || got != expected {
			t.Errorf("ParseEuros(%q) = %d, %v ; attendu %d", input, got, err, expected)
		}
	}

	for _, input := range []string{"", "abc", "19,999", "-5", "+5", "19,-5", "19,", ",50", "1,2,3", "99999999999999999999"} {
		if _, err := ParseEuros(input); err == nil {
			t.Errorf("ParseEuros(%q) aurait dû échouer", input)
		}
	}
}

func TestFormat(t *testing.T) {
	if got := Format(1999); got != "19,99 €" {
		t.Errorf("Format(1999) = %q", got)
	}
	if got := Format(5); got != "0,05 €" {
		t.Errorf("Format(5) = %q", got)
	}
}
