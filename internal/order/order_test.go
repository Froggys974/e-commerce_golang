package order

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	createdItems []ItemInput
	updateCalls  int
}

func (fake *fakeRepository) CreateFromCart(ctx context.Context, userID int64, orderRef, cardLast4 string) error {
	return nil
}

func (fake *fakeRepository) Create(ctx context.Context, userID int64, orderRef string, items []ItemInput) error {
	fake.createdItems = items
	return nil
}

func (fake *fakeRepository) FindByRef(ctx context.Context, orderRef string) (Order, error) {
	return Order{Ref: orderRef}, nil
}

func (fake *fakeRepository) ListByUser(ctx context.Context, userID int64) ([]Order, error) {
	return nil, nil
}

func (fake *fakeRepository) ListAll(ctx context.Context) ([]Order, error) {
	return nil, nil
}

func (fake *fakeRepository) UpdateStatus(ctx context.Context, orderRef string, status Status, reason string) error {
	fake.updateCalls++
	return nil
}

func TestValidateCard(t *testing.T) {
	now := time.Date(2026, time.October, 15, 0, 0, 0, 0, time.UTC)
	valid := Card{Number: "4242 4242 4242 4242", ExpMonth: 12, ExpYear: 2027, CVC: "123"}

	last4, err := validateCard(valid, now)
	if err != nil || last4 != "4242" {
		t.Fatalf("validateCard = %q, %v; attendu 4242 sans erreur", last4, err)
	}

	testCases := []struct {
		name      string
		card      Card
		wantField string
	}{
		{"numéro trop court", Card{Number: "4242", ExpMonth: 12, ExpYear: 2027, CVC: "123"}, "card_number"},
		{"numéro avec lettres", Card{Number: "4242 4242 4242 42AB", ExpMonth: 12, ExpYear: 2027, CVC: "123"}, "card_number"},
		{"mois invalide", Card{Number: "4242424242424242", ExpMonth: 13, ExpYear: 2027, CVC: "123"}, "exp_month"},
		{"carte expirée", Card{Number: "4242424242424242", ExpMonth: 9, ExpYear: 2026, CVC: "123"}, "exp_year"},
		{"cvc invalide", Card{Number: "4242424242424242", ExpMonth: 12, ExpYear: 2027, CVC: "12"}, "cvc"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := validateCard(testCase.card, now)

			var validationErr ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("erreur = %v; attendu une ValidationError", err)
			}
			if validationErr.Field != testCase.wantField {
				t.Errorf("Field = %q; attendu %q", validationErr.Field, testCase.wantField)
			}
		})
	}
}

func TestUpdateStatusValidation(t *testing.T) {
	testCases := []struct {
		name      string
		status    Status
		reason    string
		wantField string
	}{
		{"statut inconnu", Status("lost"), "", "status"},
		{"annulation sans raison", StatusCancelled, "  ", "reason"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repo := &fakeRepository{}

			_, err := NewService(repo).UpdateStatus(context.Background(), "CMD-AAAAAA", testCase.status, testCase.reason)

			var validationErr ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("erreur = %v; attendu une ValidationError", err)
			}
			if validationErr.Field != testCase.wantField {
				t.Errorf("Field = %q; attendu %q", validationErr.Field, testCase.wantField)
			}
			if repo.updateCalls != 0 {
				t.Error("la commande ne doit pas être modifiée")
			}
		})
	}
}

func TestCreateForUserMergesDuplicates(t *testing.T) {
	repo := &fakeRepository{}

	_, err := NewService(repo).CreateForUser(context.Background(), 1, []ItemInput{
		{ProductRef: "pdt-aaaaaa", Quantity: 1},
		{ProductRef: "PDT-AAAAAA", Quantity: 2},
	})
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if len(repo.createdItems) != 1 || repo.createdItems[0].Quantity != 3 {
		t.Errorf("items = %+v; attendu un seul PDT-AAAAAA en quantité 3", repo.createdItems)
	}

	_, err = NewService(&fakeRepository{}).CreateForUser(context.Background(), 1, nil)
	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("sans produit : erreur = %v; attendu une ValidationError", err)
	}
}
