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

func TestCode(t *testing.T) {
	testCases := []struct {
		length int
		want   string
	}{
		{1, `^[A-HJ-NP-Z2-9]{1}$`},
		{8, `^[A-HJ-NP-Z2-9]{8}$`},
		{0, `^$`},
	}

	for _, testCase := range testCases {
		pattern := regexp.MustCompile(testCase.want)
		for range 1000 {
			generated := Code(testCase.length)
			if !pattern.MatchString(generated) {
				t.Fatalf("Code(%d) = %q, format invalide", testCase.length, generated)
			}
		}
	}
}
