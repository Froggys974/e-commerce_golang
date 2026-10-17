package cart

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Froggys974/e-commerce_golang/internal/db"
	"github.com/Froggys974/e-commerce_golang/internal/product"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// cree le panier au premier appel puis renvoie toujours le meme
func (repo *Repository) GetOrCreate(ctx context.Context, userID int64, newRef string) (int64, string, error) {
	var cartID int64
	var cartRef string
	err := repo.db.QueryRowContext(ctx,
		`INSERT INTO carts (ref, user_id) VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET updated_at = now()
		RETURNING id, ref`,
		newRef, userID,
	).Scan(&cartID, &cartRef)
	if db.IsUniqueViolation(err) {
		return 0, "", ErrRefTaken
	}
	if err != nil {
		return 0, "", fmt.Errorf("récupération du panier : %w", err)
	}
	return cartID, cartRef, nil
}

func (repo *Repository) Items(ctx context.Context, cartID int64) ([]Item, error) {
	rows, err := repo.db.QueryContext(ctx,
		`SELECT p.ref, p.name, ci.quantity, p.price_ht_cents, p.vat_percent
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY p.name`,
		cartID,
	)
	if err != nil {
		return nil, fmt.Errorf("lecture du panier : %w", err)
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var item Item
		var priceHTCents int64
		var vatPercent int
		if err := rows.Scan(&item.ProductRef, &item.Name, &item.Quantity, &priceHTCents, &vatPercent); err != nil {
			return nil, fmt.Errorf("lecture d'une ligne du panier : %w", err)
		}
		item.UnitPriceTTCCents = product.PriceTTC(priceHTCents, vatPercent)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("lecture du panier : %w", err)
	}
	return items, nil
}

func (repo *Repository) SetItem(ctx context.Context, cartID int64, productRef string, quantity int) error {
	result, err := repo.db.ExecContext(ctx,
		`INSERT INTO cart_items (cart_id, product_id, quantity)
		SELECT $1, id, $3 FROM products WHERE ref = $2
		ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = EXCLUDED.quantity`,
		cartID, productRef, quantity,
	)
	if err != nil {
		return fmt.Errorf("ajout au panier : %w", err)
	}
	return checkFound(result, ErrProductNotFound)
}

func (repo *Repository) RemoveItem(ctx context.Context, cartID int64, productRef string) error {
	result, err := repo.db.ExecContext(ctx,
		`DELETE FROM cart_items
		WHERE cart_id = $1 AND product_id = (SELECT id FROM products WHERE ref = $2)`,
		cartID, productRef,
	)
	if err != nil {
		return fmt.Errorf("retrait du panier : %w", err)
	}
	return checkFound(result, ErrItemNotFound)
}

func checkFound(result sql.Result, notFound error) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("lignes modifiées : %w", err)
	}
	if rowsAffected == 0 {
		return notFound
	}
	return nil
}
