package ref

import (
	"regexp"
	"testing"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		prefix string
		want   string
	}{
		{Product, `^PDT-[A-HJ-NP-Z2-9]{6}$`},
		{Cart, `^BSK-[A-HJ-NP-Z2-9]{6}$`},
		{Order, `^CMD-[A-HJ-NP-Z2-9]{6}$`},
	}

	for _, testCase := range testCases {
		pattern := regexp.MustCompile(testCase.want)
		for range 1000 {
			generated := New(testCase.prefix)
			if !pattern.MatchString(generated) {
				t.Fatalf("New(%q) = %q, format invalide", testCase.prefix, generated)
			}
		}
	}
}
