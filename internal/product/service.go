package product

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Froggys974/e-commerce_golang/internal/ref"
)

type repository interface {
	Create(ctx context.Context, product *Product) error
	List(ctx context.Context, filters SearchFilters) ([]Product, error)
	FindByRef(ctx context.Context, productRef string) (Product, error)
}

const (
	defaultVATPercent = 20
	maxVATPercent     = 100
	maxPriceHTCents   = 10_000_000_000 // evite le depassement dans PriceTTC
	maxRefTries       = 5
)

type CreateInput struct {
	Name         string
	Description  string
	Category     string
	PriceHTCents int64
	VATPercent   *int
}

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Product, error) {
	product := Product{
		Name:         strings.TrimSpace(input.Name),
		Description:  strings.TrimSpace(input.Description),
		Category:     strings.ToLower(strings.TrimSpace(input.Category)),
		PriceHTCents: input.PriceHTCents,
		VATPercent:   defaultVATPercent,
	}
	if input.VATPercent != nil {
		product.VATPercent = *input.VATPercent
	}
	if err := validate(product); err != nil {
		return Product{}, err
	}
	product.PriceTTCCents = PriceTTC(product.PriceHTCents, product.VATPercent)

	for range maxRefTries {
		product.Ref = ref.New(ref.Product)
		err := s.repo.Create(ctx, &product)
		if errors.Is(err, ErrRefTaken) {
			continue
		}
		if err != nil {
			return Product{}, err
		}
		return product, nil
	}
	return Product{}, fmt.Errorf("création du produit : %d collisions de référence d'affilée", maxRefTries)
}

func validate(product Product) error {
	switch {
	case product.Name == "":
		return ValidationError{Field: "name", Message: "obligatoire"}
	case product.Category == "":
		return ValidationError{Field: "category", Message: "obligatoire"}
	case product.PriceHTCents < 0 || product.PriceHTCents > maxPriceHTCents:
		return ValidationError{Field: "price_ht_cents", Message: "doit être entre 0 et 10 000 000 000"}
	case product.VATPercent < 0 || product.VATPercent > maxVATPercent:
		return ValidationError{Field: "vat_percent", Message: "doit être entre 0 et 100"}
	}
	return nil
}

func (s *Service) Search(ctx context.Context, filters SearchFilters) ([]Product, error) {
	filters.Name = strings.TrimSpace(filters.Name)
	filters.Description = strings.TrimSpace(filters.Description)
	filters.Category = strings.ToLower(strings.TrimSpace(filters.Category))

	if err := validateSearch(filters); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, filters)
}

func (s *Service) GetByRef(ctx context.Context, productRef string) (Product, error) {
	productRef = strings.ToUpper(strings.TrimSpace(productRef))
	if productRef == "" {
		return Product{}, ErrNotFound
	}
	return s.repo.FindByRef(ctx, productRef)
}

func validateSearch(filters SearchFilters) error {
	priceFilters := []struct {
		field string
		value *int64
	}{
		{"min_price_ht", filters.MinPriceHT},
		{"max_price_ht", filters.MaxPriceHT},
		{"min_price_ttc", filters.MinPriceTTC},
		{"max_price_ttc", filters.MaxPriceTTC},
	}
	for _, priceFilter := range priceFilters {
		if priceFilter.value != nil && *priceFilter.value < 0 {
			return ValidationError{Field: priceFilter.field, Message: "doit être >= 0"}
		}
	}

	if isInverted(filters.MinPriceHT, filters.MaxPriceHT) {
		return ValidationError{Field: "min_price_ht", Message: "doit être <= max_price_ht"}
	}
	if isInverted(filters.MinPriceTTC, filters.MaxPriceTTC) {
		return ValidationError{Field: "min_price_ttc", Message: "doit être <= max_price_ttc"}
	}
	return nil
}

func isInverted(minimum, maximum *int64) bool {
	return minimum != nil && maximum != nil && *minimum > *maximum
}
