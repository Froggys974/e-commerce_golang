package product

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuildSearchQueryWithoutFilters(t *testing.T) {
	query, args := buildSearchQuery(SearchFilters{})

	if strings.Contains(query, "WHERE") {
		t.Errorf("query = %q; pas de WHERE attendu sans filtre", query)
	}
	if len(args) != 0 {
		t.Errorf("args = %v; attendu vide", args)
	}
}

func TestBuildSearchQueryWithAllFilters(t *testing.T) {
	cents := func(value int64) *int64 { return &value }

	query, args := buildSearchQuery(SearchFilters{
		Name:        "clavier",
		Description: "brown",
		Category:    "informatique",
		MinPriceHT:  cents(100),
		MaxPriceHT:  cents(200),
		MinPriceTTC: cents(120),
		MaxPriceTTC: cents(240),
	})

	wantConditions := []string{
		"name ILIKE $1",
		"description ILIKE $2",
		"category = $3",
		"price_ht_cents >= $4",
		"price_ht_cents <= $5",
		priceTTCSQL + " >= $6",
		priceTTCSQL + " <= $7",
	}
	for _, condition := range wantConditions {
		if !strings.Contains(query, condition) {
			t.Errorf("query = %q; condition %q manquante", query, condition)
		}
	}

	wantArgs := []any{"%clavier%", "%brown%", "informatique", int64(100), int64(200), int64(120), int64(240)}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v; attendu %v", args, wantArgs)
	}
}

func TestBuildSearchQueryInjection(t *testing.T) {
	injection := "'; DROP TABLE users;--"
	query, args := buildSearchQuery(SearchFilters{Name: injection})

	if strings.Contains(query, "DROP") {
		t.Errorf("query = %q; la valeur utilisateur ne doit jamais être dans le SQL", query)
	}
	if len(args) != 1 || args[0] != "%"+injection+"%" {
		t.Errorf("args = %v; la valeur doit être passée en paramètre", args)
	}
}

func TestContainsPattern(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{"clavier", "%clavier%"},
		{"50%", `%50\%%`},
		{"a_b", `%a\_b%`},
		{`c:\dir`, `%c:\\dir%`},
	}

	for _, testCase := range testCases {
		got := containsPattern(testCase.input)
		if got != testCase.want {
			t.Errorf("containsPattern(%q) = %q, want %q", testCase.input, got, testCase.want)
		}
	}
}
