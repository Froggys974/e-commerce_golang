package product

import (
	"context"
	"errors"
	"regexp"
	"testing"
)

func TestPriceTTC(t *testing.T) {
	testCases := []struct {
		priceHTCents int64
		vatPercent   int
		want         int64
	}{
		{1000, 20, 1200},
		{999, 20, 1199},
		{3, 20, 4},
		{1, 20, 1},
		{1005, 10, 1106},
		{1000, 0, 1000},
		{0, 20, 0},
	}

	for _, testCase := range testCases {
		got := PriceTTC(testCase.priceHTCents, testCase.vatPercent)
		if got != testCase.want {
			t.Errorf("PriceTTC(%d, %d) = %d, want %d",
				testCase.priceHTCents, testCase.vatPercent, got, testCase.want)
		}
	}
}

type fakeRepository struct {
	collisions int
	created    []Product
}

func (f *fakeRepository) Create(ctx context.Context, p *Product) error {
	if f.collisions > 0 {
		f.collisions--
		return ErrRefTaken
	}
	p.ID = int64(len(f.created) + 1)
	f.created = append(f.created, *p)
	return nil
}

func TestCreate(t *testing.T) {
	repo := &fakeRepository{collisions: 2}
	svc := NewService(repo)

	p, err := svc.Create(context.Background(), CreateInput{
		Name:         "  Clavier  ",
		Category:     " Informatique ",
		PriceHTCents: 1000,
	})
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}

	if !regexp.MustCompile(`^PDT-[A-HJ-NP-Z2-9]{6}$`).MatchString(p.Ref) {
		t.Errorf("Ref = %q; format invalide", p.Ref)
	}
	if p.Name != "Clavier" || p.Category != "informatique" {
		t.Errorf("Name = %q, Category = %q; attendu nettoyés", p.Name, p.Category)
	}
	if p.VATPercent != 20 || p.PriceTTCCents != 1200 {
		t.Errorf("VATPercent = %d, PriceTTCCents = %d; attendu 20 et 1200", p.VATPercent, p.PriceTTCCents)
	}
	if len(repo.created) != 1 {
		t.Errorf("len(created) = %d; attendu 1 après 2 collisions", len(repo.created))
	}
}

func TestCreateValidation(t *testing.T) {
	vat := func(v int) *int { return &v }

	tests := []struct {
		name      string
		input     CreateInput
		wantField string
	}{
		{"nom vide", CreateInput{Name: " ", Category: "info", PriceHTCents: 1000}, "name"},
		{"catégorie vide", CreateInput{Name: "Clavier", PriceHTCents: 1000}, "category"},
		{"prix négatif", CreateInput{Name: "Clavier", Category: "info", PriceHTCents: -1}, "price_ht_cents"},
		{"prix trop grand", CreateInput{Name: "Clavier", Category: "info", PriceHTCents: maxPriceHTCents + 1}, "price_ht_cents"},
		{"tva invalide", CreateInput{Name: "Clavier", Category: "info", PriceHTCents: 1000, VATPercent: vat(150)}, "vat_percent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{}
			_, err := NewService(repo).Create(context.Background(), tt.input)

			var valErr ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("erreur = %v; attendu une ValidationError", err)
			}
			if valErr.Field != tt.wantField {
				t.Errorf("Field = %q; attendu %q", valErr.Field, tt.wantField)
			}
			if len(repo.created) != 0 {
				t.Error("aucun produit ne devrait être créé")
			}
		})
	}
}
