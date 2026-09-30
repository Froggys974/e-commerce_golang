package db

import (
    "fmt"
    "database/sql"
    _ "embed"
)

//go:embed schema.sql
var schema string

func Migrate(database *sql.DB) error {
    if _, err := database.Exec(schema); err != nil {
        return fmt.Errorf("migrate: %w", err)
    }
    return nil
}