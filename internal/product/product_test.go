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
	collisions  int
	created     []Product
	lastFilters SearchFilters // filtres recus par List, pour verifier la normalisation
	listCalls   int
}

func (f *fakeRepository) Create(ctx context.Context, product *Product) error {
	if f.collisions > 0 {
		f.collisions--
		return ErrRefTaken
	}
	product.ID = int64(len(f.created) + 1)
	f.created = append(f.created, *product)
	return nil
}

func (f *fakeRepository) List(ctx context.Context, filters SearchFilters) ([]Product, error) {
	f.lastFilters = filters
	f.listCalls++
	return f.created, nil
}

func (f *fakeRepository) FindByRef(ctx context.Context, productRef string) (Product, error) {
	for _, product := range f.created {
		if product.Ref == productRef {
			return product, nil
		}
	}
	return Product{}, ErrNotFound
}

func TestCreate(t *testing.T) {
	repo := &fakeRepository{collisions: 2}
	svc := NewService(repo)

	product, err := svc.Create(context.Background(), CreateInput{
		Name:         "  Clavier  ",
		Category:     " Informatique ",
		PriceHTCents: 1000,
	})
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}

	if !regexp.MustCompile(`^PDT-[A-HJ-NP-Z2-9]{6}$`).MatchString(product.Ref) {
		t.Errorf("Ref = %q; format invalide", product.Ref)
	}
	if product.Name != "Clavier" || product.Category != "informatique" {
		t.Errorf("Name = %q, Category = %q; attendu nettoyés", product.Name, product.Category)
	}
	if product.VATPercent != 20 || product.PriceTTCCents != 1200 {
		t.Errorf("VATPercent = %d, PriceTTCCents = %d; attendu 20 et 1200", product.VATPercent, product.PriceTTCCents)
	}
	if len(repo.created) != 1 {
		t.Errorf("len(created) = %d; attendu 1 après 2 collisions", len(repo.created))
	}
}

func TestCreateValidation(t *testing.T) {
	vat := func(value int) *int { return &value }

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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			repo := &fakeRepository{}
			_, err := NewService(repo).Create(context.Background(), testCase.input)

			var valErr ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("erreur = %v; attendu une ValidationError", err)
			}
			if valErr.Field != testCase.wantField {
				t.Errorf("Field = %q; attendu %q", valErr.Field, testCase.wantField)
			}
			if len(repo.created) != 0 {
				t.Error("aucun produit ne devrait être créé")
			}
		})
	}
}

func TestSearchNormalizesFilters(t *testing.T) {
	repo := &fakeRepository{}
	_, err := NewService(repo).Search(context.Background(), SearchFilters{
		Name:     "  clavier ",
		Category: " Informatique ",
	})
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if repo.lastFilters.Name != "clavier" || repo.lastFilters.Category != "informatique" {
		t.Errorf("filtres = %+v; attendu name et category nettoyés", repo.lastFilters)
	}
}

func TestSearchValidation(t *testing.T) {
	cents := func(value int64) *int64 { return &value }

	testCases := []struct {
		name      string
		filters   SearchFilters
		wantField string
	}{
		{"min HT négatif", SearchFilters{MinPriceHT: cents(-1)}, "min_price_ht"},
		{"max TTC négatif", SearchFilters{MaxPriceTTC: cents(-5)}, "max_price_ttc"},
		{"HT min > max", SearchFilters{MinPriceHT: cents(2000), MaxPriceHT: cents(1000)}, "min_price_ht"},
		{"TTC min > max", SearchFilters{MinPriceTTC: cents(2000), MaxPriceTTC: cents(1000)}, "min_price_ttc"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repo := &fakeRepository{}
			_, err := NewService(repo).Search(context.Background(), testCase.filters)

			var validationErr ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("erreur = %v; attendu une ValidationError", err)
			}
			if validationErr.Field != testCase.wantField {
				t.Errorf("Field = %q; attendu %q", validationErr.Field, testCase.wantField)
			}
			if repo.listCalls != 0 {
				t.Error("la base ne doit pas être interrogée si les filtres sont invalides")
			}
		})
	}
}

func TestGetByRef(t *testing.T) {
	repo := &fakeRepository{created: []Product{{Ref: "PDT-TS2N8E", Name: "Clavier"}}}
	service := NewService(repo)

	product, err := service.GetByRef(context.Background(), "  pdt-ts2n8e ")
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if product.Name != "Clavier" {
		t.Errorf("Name = %q; attendu Clavier", product.Name)
	}

	for _, missingRef := range []string{"PDT-AAAAAA", "   "} {
		_, err = service.GetByRef(context.Background(), missingRef)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("GetByRef(%q) erreur = %v; attendu ErrNotFound", missingRef, err)
		}
	}
}
