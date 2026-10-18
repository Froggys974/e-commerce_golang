package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Froggys974/e-commerce_golang/internal/db"
	"github.com/Froggys974/e-commerce_golang/internal/product"
)

const selectOrders = `SELECT o.id, o.ref, o.user_id, u.email, o.status, COALESCE(o.cancel_reason, ''), o.created_at,
	p.ref, p.name, oi.quantity, oi.unit_price_ttc_cents
FROM orders o
JOIN users u ON u.id = o.user_id
JOIN order_items oi ON oi.order_id = o.id
JOIN products p ON p.id = oi.product_id`

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type cartLine struct {
	productID         int64
	quantity          int
	unitPriceTTCCents int64
}

func (repo *Repository) CreateFromCart(ctx context.Context, userID int64, orderRef, cardLast4 string) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("paiement : %w", err)
	}
	// sans effet apres commit
	defer tx.Rollback()

	lines, err := readCart(ctx, tx, userID)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return ErrEmptyCart
	}

	orderID, err := insertOrder(ctx, tx, orderRef, userID, StatusPaid)
	if err != nil {
		return err
	}

	var totalCents int64
	for _, line := range lines {
		if err := insertItem(ctx, tx, orderID, line); err != nil {
			return err
		}
		totalCents += line.unitPriceTTCCents * int64(line.quantity)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO payments (order_id, card_last4, amount_cents) VALUES ($1, $2, $3)`,
		orderID, cardLast4, totalCents,
	)
	if err != nil {
		return fmt.Errorf("enregistrement du paiement : %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`DELETE FROM cart_items WHERE cart_id = (SELECT id FROM carts WHERE user_id = $1)`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("vidage du panier : %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("paiement : %w", err)
	}
	return nil
}

func (repo *Repository) Create(ctx context.Context, userID int64, orderRef string, items []ItemInput) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("création de la commande : %w", err)
	}
	// sans effet apres commit
	defer tx.Rollback()

	var userExists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&userExists)
	if err != nil {
		return fmt.Errorf("création de la commande : %w", err)
	}
	if !userExists {
		return ErrUserNotFound
	}

	orderID, err := insertOrder(ctx, tx, orderRef, userID, StatusPending)
	if err != nil {
		return err
	}

	for _, item := range items {
		line := cartLine{quantity: item.Quantity}
		var priceHTCents int64
		var vatPercent int
		err := tx.QueryRowContext(ctx,
			`SELECT id, price_ht_cents, vat_percent FROM products WHERE ref = $1`,
			item.ProductRef,
		).Scan(&line.productID, &priceHTCents, &vatPercent)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrProductNotFound
		}
		if err != nil {
			return fmt.Errorf("lecture du produit %s : %w", item.ProductRef, err)
		}
		line.unitPriceTTCCents = product.PriceTTC(priceHTCents, vatPercent)

		if err := insertItem(ctx, tx, orderID, line); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("création de la commande : %w", err)
	}
	return nil
}

func (repo *Repository) FindByRef(ctx context.Context, orderRef string) (Order, error) {
	orders, err := repo.list(ctx, selectOrders+" WHERE o.ref = $1", orderRef)
	if err != nil {
		return Order{}, err
	}
	if len(orders) == 0 {
		return Order{}, ErrNotFound
	}
	return orders[0], nil
}

func (repo *Repository) ListByUser(ctx context.Context, userID int64) ([]Order, error) {
	return repo.list(ctx, selectOrders+" WHERE o.user_id = $1 ORDER BY o.created_at DESC, o.id DESC, p.name", userID)
}

func (repo *Repository) ListAll(ctx context.Context) ([]Order, error) {
	return repo.list(ctx, selectOrders+" ORDER BY o.created_at DESC, o.id DESC, p.name")
}

func (repo *Repository) UpdateStatus(ctx context.Context, orderRef string, previous, status Status, reason string) error {
	result, err := repo.db.ExecContext(ctx,
		`UPDATE orders SET status = $2, cancel_reason = NULLIF($3, '') WHERE ref = $1 AND status = $4`,
		orderRef, status, reason, previous,
	)
	if err != nil {
		return fmt.Errorf("changement de statut : %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("changement de statut : %w", err)
	}
	if rowsAffected == 0 {
		return ErrStatusChanged
	}
	return nil
}

// une ligne sql par produit commande regroupees par commande
// any impose par QueryContext
func (repo *Repository) list(ctx context.Context, query string, args ...any) ([]Order, error) {
	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("lecture des commandes : %w", err)
	}
	defer rows.Close()

	orders := []Order{}
	positions := map[int64]int{}
	for rows.Next() {
		var orderID int64
		var current Order
		var item Item
		err := rows.Scan(
			&orderID, &current.Ref, &current.UserID, &current.UserEmail, &current.Status, &current.CancelReason, &current.CreatedAt,
			&item.ProductRef, &item.ProductName, &item.Quantity, &item.UnitPriceTTCCents,
		)
		if err != nil {
			return nil, fmt.Errorf("lecture d'une commande : %w", err)
		}

		position, known := positions[orderID]
		if !known {
			orders = append(orders, current)
			position = len(orders) - 1
			positions[orderID] = position
		}
		orders[position].Items = append(orders[position].Items, item)
		orders[position].TotalTTCCents += item.UnitPriceTTCCents * int64(item.Quantity)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("lecture des commandes : %w", err)
	}
	return orders, nil
}

func readCart(ctx context.Context, tx *sql.Tx, userID int64) ([]cartLine, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT ci.product_id, ci.quantity, p.price_ht_cents, p.vat_percent
		FROM carts c
		JOIN cart_items ci ON ci.cart_id = c.id
		JOIN products p ON p.id = ci.product_id
		WHERE c.user_id = $1
		FOR UPDATE OF ci`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("lecture du panier : %w", err)
	}
	defer rows.Close()

	var lines []cartLine
	for rows.Next() {
		var line cartLine
		var priceHTCents int64
		var vatPercent int
		if err := rows.Scan(&line.productID, &line.quantity, &priceHTCents, &vatPercent); err != nil {
			return nil, fmt.Errorf("lecture du panier : %w", err)
		}
		line.unitPriceTTCCents = product.PriceTTC(priceHTCents, vatPercent)
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("lecture du panier : %w", err)
	}
	return lines, nil
}

func insertOrder(ctx context.Context, tx *sql.Tx, orderRef string, userID int64, status Status) (int64, error) {
	var orderID int64
	err := tx.QueryRowContext(ctx,
		`INSERT INTO orders (ref, user_id, status) VALUES ($1, $2, $3) RETURNING id`,
		orderRef, userID, status,
	).Scan(&orderID)
	if db.IsUniqueViolation(err) {
		return 0, ErrRefTaken
	}
	if err != nil {
		return 0, fmt.Errorf("création de la commande : %w", err)
	}
	return orderID, nil
}

func insertItem(ctx context.Context, tx *sql.Tx, orderID int64, line cartLine) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO order_items (order_id, product_id, quantity, unit_price_ttc_cents) VALUES ($1, $2, $3, $4)`,
		orderID, line.productID, line.quantity, line.unitPriceTTCCents,
	)
	if err != nil {
		return fmt.Errorf("ajout d'un produit à la commande : %w", err)
	}
	return nil
}
