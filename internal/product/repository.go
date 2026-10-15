package product

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Froggys974/e-commerce_golang/internal/db"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, p *Product) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO products (ref, name, description, category, price_ht_cents, vat_percent)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`,
		p.Ref, p.Name, p.Description, p.Category, p.PriceHTCents, p.VATPercent,
	).Scan(&p.ID, &p.CreatedAt)
	if db.IsUniqueViolation(err) {
		return ErrRefTaken
	}
	if err != nil {
		return fmt.Errorf("création du produit : %w", err)
	}
	return nil
}
