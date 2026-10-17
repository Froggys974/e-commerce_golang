package cart

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Froggys974/e-commerce_golang/internal/ref"
)

type repository interface {
	GetOrCreate(ctx context.Context, userID int64, newRef string) (int64, string, error)
	Items(ctx context.Context, cartID int64) ([]Item, error)
	SetItem(ctx context.Context, cartID int64, productRef string, quantity int) error
	RemoveItem(ctx context.Context, cartID int64, productRef string) error
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

func (service *Service) Get(ctx context.Context, userID int64) (Cart, error) {
	cartID, cartRef, err := service.cartOf(ctx, userID)
	if err != nil {
		return Cart{}, err
	}
	return service.load(ctx, cartID, cartRef)
}

func (service *Service) SetItem(ctx context.Context, userID int64, productRef string, quantity int) (Cart, error) {
	if quantity < minQuantity || quantity > maxQuantity {
		return Cart{}, ValidationError{Field: "quantity", Message: "doit être entre 1 et 99"}
	}

	cartID, cartRef, err := service.cartOf(ctx, userID)
	if err != nil {
		return Cart{}, err
	}
	if err := service.repo.SetItem(ctx, cartID, normalizeRef(productRef), quantity); err != nil {
		return Cart{}, err
	}
	return service.load(ctx, cartID, cartRef)
}

func (service *Service) RemoveItem(ctx context.Context, userID int64, productRef string) (Cart, error) {
	cartID, cartRef, err := service.cartOf(ctx, userID)
	if err != nil {
		return Cart{}, err
	}
	if err := service.repo.RemoveItem(ctx, cartID, normalizeRef(productRef)); err != nil {
		return Cart{}, err
	}
	return service.load(ctx, cartID, cartRef)
}

func (service *Service) cartOf(ctx context.Context, userID int64) (int64, string, error) {
	for range maxRefTries {
		cartID, cartRef, err := service.repo.GetOrCreate(ctx, userID, ref.New(ref.Cart))
		if errors.Is(err, ErrRefTaken) {
			continue
		}
		return cartID, cartRef, err
	}
	return 0, "", fmt.Errorf("panier : %d collisions de référence d'affilée", maxRefTries)
}

func (service *Service) load(ctx context.Context, cartID int64, cartRef string) (Cart, error) {
	items, err := service.repo.Items(ctx, cartID)
	if err != nil {
		return Cart{}, err
	}

	cart := Cart{Ref: cartRef, Items: items}
	for index := range cart.Items {
		item := &cart.Items[index]
		item.LineTotalCents = item.UnitPriceTTCCents * int64(item.Quantity)
		cart.TotalTTCCents += item.LineTotalCents
	}
	return cart, nil
}

func normalizeRef(productRef string) string {
	return strings.ToUpper(strings.TrimSpace(productRef))
}
