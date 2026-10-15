package product

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Froggys974/e-commerce_golang/internal/ref"
)

type repository interface {
	Create(ctx context.Context, p *Product) error
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

func NewService(r repository) *Service {
	return &Service{repo: r}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Product, error) {
	p := Product{
		Name:         strings.TrimSpace(input.Name),
		Description:  strings.TrimSpace(input.Description),
		Category:     strings.ToLower(strings.TrimSpace(input.Category)),
		PriceHTCents: input.PriceHTCents,
		VATPercent:   defaultVATPercent,
	}
	if input.VATPercent != nil {
		p.VATPercent = *input.VATPercent
	}
	if err := validate(p); err != nil {
		return Product{}, err
	}
	p.PriceTTCCents = PriceTTC(p.PriceHTCents, p.VATPercent)

	for range maxRefTries {
		p.Ref = ref.New(ref.Product)
		err := s.repo.Create(ctx, &p)
		if errors.Is(err, ErrRefTaken) {
			continue
		}
		if err != nil {
			return Product{}, err
		}
		return p, nil
	}
	return Product{}, fmt.Errorf("création du produit : %d collisions de référence d'affilée", maxRefTries)
}

func validate(p Product) error {
	switch {
	case p.Name == "":
		return ValidationError{Field: "name", Message: "obligatoire"}
	case p.Category == "":
		return ValidationError{Field: "category", Message: "obligatoire"}
	case p.PriceHTCents < 0 || p.PriceHTCents > maxPriceHTCents:
		return ValidationError{Field: "price_ht_cents", Message: "doit être entre 0 et 10 000 000 000"}
	case p.VATPercent < 0 || p.VATPercent > maxVATPercent:
		return ValidationError{Field: "vat_percent", Message: "doit être entre 0 et 100"}
	}
	return nil
}
