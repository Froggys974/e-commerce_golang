package client

import (
	"testing"

	"github.com/Froggys974/e-commerce_golang/internal/product"
)

func cents(value int64) *int64 {
	return &value
}

func TestParseSearch(t *testing.T) {
	testCases := []struct {
		input string
		want  product.SearchFilters
	}{
		{"", product.SearchFilters{}},
		{"clavier", product.SearchFilters{Name: "clavier"}},
		{"sans fil", product.SearchFilters{Name: "sans fil"}},
		{"cat:livres", product.SearchFilters{Category: "livres"}},
		{"desc:bio desc:coton", product.SearchFilters{Description: "bio coton"}},
		{"clavier cat:informatique max:100", product.SearchFilters{Name: "clavier", Category: "informatique", MaxPriceTTC: cents(10000)}},
		{"min:9,99 maxht:20.5", product.SearchFilters{MinPriceTTC: cents(999), MaxPriceHT: cents(2050)}},
		{"MINHT:1 CAT:jeux", product.SearchFilters{Category: "jeux", MinPriceHT: cents(100)}},
		{"souris cat:", product.SearchFilters{Name: "souris"}},
		{"usb:c", product.SearchFilters{Name: "usb:c"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			got, err := parseSearch(testCase.input)
			if err != nil {
				t.Fatalf("erreur inattendue : %v", err)
			}
			if !sameFilters(got, testCase.want) {
				t.Errorf("parseSearch(%q) = %+v; attendu %+v", testCase.input, describe(got), describe(testCase.want))
			}
		})
	}
}

func TestParseSearchInvalidPrice(t *testing.T) {
	for _, input := range []string{"max:abc", "min:19,999", "maxht:-5"} {
		if _, err := parseSearch(input); err == nil {
			t.Errorf("parseSearch(%q) aurait dû échouer", input)
		}
	}
}

func sameFilters(first, second product.SearchFilters) bool {
	return first.Name == second.Name &&
		first.Description == second.Description &&
		first.Category == second.Category &&
		samePrice(first.MinPriceHT, second.MinPriceHT) &&
		samePrice(first.MaxPriceHT, second.MaxPriceHT) &&
		samePrice(first.MinPriceTTC, second.MinPriceTTC) &&
		samePrice(first.MaxPriceTTC, second.MaxPriceTTC)
}

func samePrice(first, second *int64) bool {
	if first == nil || second == nil {
		return first == second
	}
	return *first == *second
}

func describe(filters product.SearchFilters) map[string]any {
	values := map[string]any{"name": filters.Name, "desc": filters.Description, "cat": filters.Category}
	prices := map[string]*int64{"minht": filters.MinPriceHT, "maxht": filters.MaxPriceHT, "min": filters.MinPriceTTC, "max": filters.MaxPriceTTC}
	for key, value := range prices {
		if value != nil {
			values[key] = *value
		}
	}
	return values
}
