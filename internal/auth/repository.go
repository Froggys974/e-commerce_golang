package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) CreateToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := repo.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		tokenHash, userID, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("création du refresh token : %w", err)
	}
	return nil
}

func (repo *Repository) RotateToken(ctx context.Context, oldHash, newHash string, expiresAt time.Time) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("rotation du refresh token : %w", err)
	}
	// sans effet apres commit
	defer tx.Rollback()

	var userID int64
	err = tx.QueryRowContext(ctx,
		`DELETE FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > now()
		RETURNING user_id`,
		oldHash,
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrInvalidToken
	}
	if err != nil {
		return 0, fmt.Errorf("rotation du refresh token : %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO refresh_tokens (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		newHash, userID, expiresAt,
	)
	if err != nil {
		return 0, fmt.Errorf("rotation du refresh token : %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("rotation du refresh token : %w", err)
	}
	return userID, nil
}

func (repo *Repository) RevokeToken(ctx context.Context, tokenHash string) error {
	_, err := repo.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("suppression du refresh token : %w", err)
	}
	return nil
}
