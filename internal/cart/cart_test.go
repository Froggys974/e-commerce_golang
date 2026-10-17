package cart

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	items    []Item
	setCalls int
}

func (fake *fakeRepository) GetOrCreate(ctx context.Context, userID int64, newRef string) (int64, string, error) {
	return 1, "BSK-AAAAAA", nil
}

func (fake *fakeRepository) Items(ctx context.Context, cartID int64) ([]Item, error) {
	return fake.items, nil
}

func (fake *fakeRepository) SetItem(ctx context.Context, cartID int64, productRef string, quantity int) error {
	fake.setCalls++
	return nil
}

func (fake *fakeRepository) RemoveItem(ctx context.Context, cartID int64, productRef string) error {
	return nil
}

func TestGetComputesTotals(t *testing.T) {
	repo := &fakeRepository{items: []Item{
		{ProductRef: "PDT-AAAAAA", Quantity: 2, UnitPriceTTCCents: 1200},
		{ProductRef: "PDT-BBBBBB", Quantity: 1, UnitPriceTTCCents: 599},
	}}

	cart, err := NewService(repo).Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}

	if cart.Items[0].LineTotalCents != 2400 {
		t.Errorf("LineTotalCents = %d; attendu 2400", cart.Items[0].LineTotalCents)
	}
	if cart.TotalTTCCents != 2999 {
		t.Errorf("TotalTTCCents = %d; attendu 2999", cart.TotalTTCCents)
	}
}

func TestSetItemQuantityValidation(t *testing.T) {
	for _, quantity := range []int{0, -1, 100} {
		repo := &fakeRepository{}

		_, err := NewService(repo).SetItem(context.Background(), 1, "PDT-AAAAAA", quantity)

		var validationErr ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("quantity %d : erreur = %v; attendu une ValidationError", quantity, err)
		}
		if repo.setCalls != 0 {
			t.Errorf("quantity %d : le panier ne doit pas être modifié", quantity)
		}
	}
}
