package db

import (
    "database/sql"
    "fmt"
    "os"

    _ "github.com/jackc/pgx/v5/stdlib"
)

func Open() (*sql.DB, error) {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "postgres://postgresuser:postgrespwd@postgres:5432/postgresdb?sslmode=disable"
    }

    database, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, fmt.Errorf("sql.Open: %w", err)
    }
    if err := database.Ping(); err != nil {
        return nil, fmt.Errorf("ping db: %w", err)
    }
    return database, nil
}
