package client

import (
	"fmt"
	"strings"

	"github.com/Froggys974/e-commerce_golang/internal/price"
	"github.com/Froggys974/e-commerce_golang/internal/product"
)

const searchSyntax = "cat: catégorie • desc: description • min: / max: prix TTC € • minht: / maxht: prix HT €"

func parseSearch(text string) (product.SearchFilters, error) {
	var filters product.SearchFilters
	var nameWords []string

	for _, word := range strings.Fields(text) {
		key, value, hasPrefix := strings.Cut(word, ":")
		if !hasPrefix {
			nameWords = append(nameWords, word)
			continue
		}
		if value == "" {
			continue
		}

		var err error
		switch strings.ToLower(key) {
		case "cat":
			filters.Category = value
		case "desc":
			filters.Description = joinWord(filters.Description, value)
		case "min":
			filters.MinPriceTTC, err = parseEurosFilter(key, value)
		case "max":
			filters.MaxPriceTTC, err = parseEurosFilter(key, value)
		case "minht":
			filters.MinPriceHT, err = parseEurosFilter(key, value)
		case "maxht":
			filters.MaxPriceHT, err = parseEurosFilter(key, value)
		default:
			nameWords = append(nameWords, word)
		}
		if err != nil {
			return product.SearchFilters{}, err
		}
	}

	filters.Name = strings.Join(nameWords, " ")
	return filters, nil
}

func parseEurosFilter(key, value string) (*int64, error) {
	cents, err := price.ParseEuros(value)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", key, err)
	}
	return &cents, nil
}

func joinWord(current, word string) string {
	if current == "" {
		return word
	}
	return current + " " + word
}
