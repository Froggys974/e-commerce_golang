package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Froggys974/e-commerce_golang/internal/db"
)

const (
	// ordre des colonnes = ordre du Scan dans scanProduct
	productColumns   = "id, ref, name, description, category, price_ht_cents, vat_percent, created_at"
	maxSearchResults = 100
	// meme calcul que PriceTTC mais fait par postgres
	priceTTCSQL = "(price_ht_cents * (100 + vat_percent) + 50) / 100"
)

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) Create(ctx context.Context, product *Product) error {
	err := repo.db.QueryRowContext(ctx,
		`INSERT INTO products (ref, name, description, category, price_ht_cents, vat_percent)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`,
		product.Ref, product.Name, product.Description, product.Category, product.PriceHTCents, product.VATPercent,
	).Scan(&product.ID, &product.CreatedAt)
	if db.IsUniqueViolation(err) {
		return ErrRefTaken
	}
	if err != nil {
		return fmt.Errorf("création du produit : %w", err)
	}
	return nil
}

func (repo *Repository) List(ctx context.Context, filters SearchFilters) ([]Product, error) {
	query, args := buildSearchQuery(filters)

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("recherche de produits : %w", err)
	}
	defer rows.Close()

	// JSON [] au lieu de null quand aucun resultat
	products := []Product{}
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("lecture d'un produit : %w", err)
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recherche de produits : %w", err)
	}
	return products, nil
}

func (repo *Repository) FindByRef(ctx context.Context, productRef string) (Product, error) {
	row := repo.db.QueryRowContext(ctx,
		"SELECT "+productColumns+" FROM products WHERE ref = $1",
		productRef,
	)
	product, err := scanProduct(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("recherche du produit %s : %w", productRef, err)
	}
	return product, nil
}

func buildSearchQuery(filters SearchFilters) (string, []any) {
	var conditions []string
	var args []any

	addCondition := func(sqlTemplate string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(sqlTemplate, len(args)))
	}

	if filters.Name != "" {
		addCondition("name ILIKE $%d", containsPattern(filters.Name))
	}
	if filters.Description != "" {
		addCondition("description ILIKE $%d", containsPattern(filters.Description))
	}
	if filters.Category != "" {
		addCondition("category = $%d", filters.Category)
	}

	priceFilters := []struct {
		value       *int64
		sqlTemplate string
	}{
		{filters.MinPriceHT, "price_ht_cents >= $%d"},
		{filters.MaxPriceHT, "price_ht_cents <= $%d"},
		{filters.MinPriceTTC, priceTTCSQL + " >= $%d"},
		{filters.MaxPriceTTC, priceTTCSQL + " <= $%d"},
	}
	for _, priceFilter := range priceFilters {
		if priceFilter.value != nil {
			addCondition(priceFilter.sqlTemplate, *priceFilter.value)
		}
	}

	query := "SELECT " + productColumns + " FROM products"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY name, id LIMIT %d", maxSearchResults)
	return query, args
}

func containsPattern(text string) string {
	return "%" + likeEscaper.Replace(text) + "%"
}

type rowScanner interface {
	Scan(destinations ...any) error
}

func scanProduct(row rowScanner) (Product, error) {
	var product Product
	err := row.Scan(
		&product.ID,
		&product.Ref,
		&product.Name,
		&product.Description,
		&product.Category,
		&product.PriceHTCents,
		&product.VATPercent,
		&product.CreatedAt,
	)
	if err != nil {
		return Product{}, err
	}
	product.PriceTTCCents = PriceTTC(product.PriceHTCents, product.VATPercent)
	return product, nil
}
