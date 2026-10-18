package db

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schema string

//go:embed seed.sql
var seed string

func Migrate(database *sql.DB) error {
	if _, err := database.Exec(schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// produits de démo : ON CONFLICT DO NOTHING, donc relançable sans doublon
func Seed(database *sql.DB) error {
	if _, err := database.Exec(seed); err != nil {
		return fmt.Errorf("seed: %w", err)
	}
	return nil
}
