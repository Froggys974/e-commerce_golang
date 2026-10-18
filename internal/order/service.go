package order

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Froggys974/e-commerce_golang/internal/ref"
)

type repository interface {
	CreateFromCart(ctx context.Context, userID int64, orderRef, cardLast4 string) error
	Create(ctx context.Context, userID int64, orderRef string, items []ItemInput) error
	FindByRef(ctx context.Context, orderRef string) (Order, error)
	ListByUser(ctx context.Context, userID int64) ([]Order, error)
	ListAll(ctx context.Context) ([]Order, error)
	UpdateStatus(ctx context.Context, orderRef string, previous, status Status, reason string) error
}

const (
	minQuantity = 1
	maxQuantity = 99
	maxRefTries = 5
)

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (service *Service) Checkout(ctx context.Context, userID int64, card Card) (Order, error) {
	cardLast4, err := validateCard(card, time.Now())
	if err != nil {
		return Order{}, err
	}
	return service.createWithRef(ctx, func(orderRef string) error {
		return service.repo.CreateFromCart(ctx, userID, orderRef, cardLast4)
	})
}

func (service *Service) CreateForUser(ctx context.Context, userID int64, items []ItemInput) (Order, error) {
	items, err := normalizeItems(items)
	if err != nil {
		return Order{}, err
	}
	return service.createWithRef(ctx, func(orderRef string) error {
		return service.repo.Create(ctx, userID, orderRef, items)
	})
}

func (service *Service) ListMine(ctx context.Context, userID int64) ([]Order, error) {
	return service.repo.ListByUser(ctx, userID)
}

func (service *Service) ListAll(ctx context.Context) ([]Order, error) {
	return service.repo.ListAll(ctx)
}

func (service *Service) UpdateStatus(ctx context.Context, orderRef string, status Status, reason string) (Order, error) {
	if !status.Valid() {
		return Order{}, ValidationError{Field: "status", Message: "doit être pending, paid, shipping, delivered ou cancelled"}
	}

	reason = strings.TrimSpace(reason)
	if status == StatusCancelled && reason == "" {
		return Order{}, ValidationError{Field: "reason", Message: "obligatoire pour une annulation"}
	}
	if status != StatusCancelled {
		reason = ""
	}

	orderRef = strings.ToUpper(strings.TrimSpace(orderRef))
	current, err := service.repo.FindByRef(ctx, orderRef)
	if err != nil {
		return Order{}, err
	}
	if !current.Status.CanBecome(status) {
		return Order{}, ValidationError{Field: "status", Message: fmt.Sprintf("passage de %s à %s interdit", current.Status, status)}
	}

	if err := service.repo.UpdateStatus(ctx, orderRef, current.Status, status, reason); err != nil {
		return Order{}, err
	}
	return service.repo.FindByRef(ctx, orderRef)
}

func (service *Service) createWithRef(ctx context.Context, create func(orderRef string) error) (Order, error) {
	for range maxRefTries {
		orderRef := ref.New(ref.Order)
		err := create(orderRef)
		if errors.Is(err, ErrRefTaken) {
			continue
		}
		if err != nil {
			return Order{}, err
		}
		return service.repo.FindByRef(ctx, orderRef)
	}
	return Order{}, fmt.Errorf("commande : %d collisions de référence d'affilée", maxRefTries)
}

// regroupe les doublons d un meme produit
func normalizeItems(items []ItemInput) ([]ItemInput, error) {
	if len(items) == 0 {
		return nil, ValidationError{Field: "items", Message: "au moins un produit"}
	}

	var normalized []ItemInput
	positions := map[string]int{}
	for _, item := range items {
		productRef := strings.ToUpper(strings.TrimSpace(item.ProductRef))
		if position, known := positions[productRef]; known {
			normalized[position].Quantity += item.Quantity
			continue
		}
		positions[productRef] = len(normalized)
		normalized = append(normalized, ItemInput{ProductRef: productRef, Quantity: item.Quantity})
	}

	for _, item := range normalized {
		if item.Quantity < minQuantity || item.Quantity > maxQuantity {
			return nil, ValidationError{Field: "quantity", Message: "doit être entre 1 et 99"}
		}
	}
	return normalized, nil
}

func validateCard(card Card, now time.Time) (string, error) {
	number := strings.ReplaceAll(card.Number, " ", "")
	if len(number) != 16 || !isDigits(number) {
		return "", ValidationError{Field: "card_number", Message: "16 chiffres"}
	}
	if !passesLuhn(number) {
		return "", ValidationError{Field: "card_number", Message: "numéro de carte invalide"}
	}
	if card.ExpMonth < 1 || card.ExpMonth > 12 {
		return "", ValidationError{Field: "exp_month", Message: "doit être entre 1 et 12"}
	}
	if card.ExpYear < now.Year() || (card.ExpYear == now.Year() && card.ExpMonth < int(now.Month())) {
		return "", ValidationError{Field: "exp_year", Message: "carte expirée"}
	}
	if len(card.CVC) != 3 || !isDigits(card.CVC) {
		return "", ValidationError{Field: "cvc", Message: "3 chiffres"}
	}
	return number[12:], nil
}

// algorithme de Luhn : en partant de la droite, on double un chiffre sur deux
// Détecte les fautes de frappe, pas les fausses cartes.
func passesLuhn(number string) bool {
	sum := 0
	double := false
	for index := len(number) - 1; index >= 0; index-- {
		digit := int(number[index] - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	return sum%10 == 0
}

func isDigits(text string) bool {
	for _, character := range text {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
