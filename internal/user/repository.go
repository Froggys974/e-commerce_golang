package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Froggys974/e-commerce_golang/internal/db"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) List(ctx context.Context) ([]User, error) {
	rows, err := repo.db.QueryContext(ctx,
		`SELECT id, email, password_hash, role, confirmed, created_at
		FROM users ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("liste des utilisateurs : %w", err)
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Confirmed, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("lecture d'un utilisateur : %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("liste des utilisateurs : %w", err)
	}
	return users, nil
}

func (repo *Repository) FindByID(ctx context.Context, id int64) (User, error) {
	var user User
	err := repo.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, role, confirmed, created_at
		FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Confirmed, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, NotFoundError{ID: id}
	}
	if err != nil {
		return User{}, fmt.Errorf("recherche de l'utilisateur %d : %w", id, err)
	}
	return user, nil
}

func (repo *Repository) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := repo.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, role, confirmed, created_at
		FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Confirmed, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrEmailNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("recherche de l'utilisateur %s : %w", email, err)
	}
	return user, nil
}

func (repo *Repository) Create(ctx context.Context, user *User) error {
	err := repo.db.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		user.Email, user.PasswordHash, user.Role,
	).Scan(&user.ID, &user.CreatedAt)
	if db.IsUniqueViolation(err) {
		return EmailTakenError{Email: user.Email}
	}
	if err != nil {
		return fmt.Errorf("création de l'utilisateur : %w", err)
	}
	return nil
}

func (repo *Repository) Update(ctx context.Context, user User) error {
	result, err := repo.db.ExecContext(ctx,
		`UPDATE users SET email = $1, role = $2 WHERE id = $3`,
		user.Email, user.Role, user.ID,
	)
	if db.IsUniqueViolation(err) {
		return EmailTakenError{Email: user.Email}
	}
	if err != nil {
		return fmt.Errorf("mise à jour de l'utilisateur %d : %w", user.ID, err)
	}
	return checkFound(result, NotFoundError{ID: user.ID})
}

func (repo *Repository) Delete(ctx context.Context, id int64) error {
	result, err := repo.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("suppression de l'utilisateur %d : %w", id, err)
	}
	return checkFound(result, NotFoundError{ID: id})
}

func (repo *Repository) Confirm(ctx context.Context, id int64) error {
	result, err := repo.db.ExecContext(ctx,
		`UPDATE users
		SET confirmed = true, confirm_code = NULL, confirm_expires_at = NULL, confirm_attempts = 0
		WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("confirmation de l'utilisateur %d : %w", id, err)
	}
	return checkFound(result, NotFoundError{ID: id})
}

func (repo *Repository) SetConfirmCode(ctx context.Context, email, code string, expiresAt time.Time) error {
	result, err := repo.db.ExecContext(ctx,
		`UPDATE users
		SET confirm_code = $2, confirm_expires_at = $3, confirm_attempts = 0
		WHERE email = $1 AND confirmed = false`,
		email, code, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("code de confirmation : %w", err)
	}
	return checkFound(result, ErrEmailNotFound)
}

func (repo *Repository) ConfirmByCode(ctx context.Context, email, code string, maxAttempts int) error {
	result, err := repo.db.ExecContext(ctx,
		`UPDATE users
		SET confirmed = true, confirm_code = NULL, confirm_expires_at = NULL, confirm_attempts = 0
		WHERE email = $1 AND confirm_code = $2 AND confirmed = false
		AND confirm_expires_at > now() AND confirm_attempts < $3`,
		email, code, maxAttempts,
	)
	if err != nil {
		return fmt.Errorf("confirmation par code : %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("confirmation par code : %w", err)
	}
	if rowsAffected == 1 {
		return nil
	}

	_, err = repo.db.ExecContext(ctx,
		`UPDATE users SET confirm_attempts = confirm_attempts + 1
		WHERE email = $1 AND confirmed = false`,
		email,
	)
	if err != nil {
		return fmt.Errorf("tentative de confirmation : %w", err)
	}
	return ErrInvalidCode
}

func (repo *Repository) SetResetCode(ctx context.Context, email, code string, expiresAt time.Time) error {
	result, err := repo.db.ExecContext(ctx,
		`UPDATE users
		SET reset_code = $2, reset_expires_at = $3, reset_attempts = 0
		WHERE email = $1`,
		email, code, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("code de réinitialisation : %w", err)
	}
	return checkFound(result, ErrEmailNotFound)
}

func (repo *Repository) ResetPassword(ctx context.Context, email, code, passwordHash string, maxAttempts int) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("réinitialisation du mot de passe : %w", err)
	}
	// sans effet apres commit
	defer tx.Rollback()

	var id int64
	err = tx.QueryRowContext(ctx,
		`UPDATE users
		SET password_hash = $3, reset_code = NULL, reset_expires_at = NULL, reset_attempts = 0
		WHERE email = $1 AND reset_code = $2
		AND reset_expires_at > now() AND reset_attempts < $4
		RETURNING id`,
		email, code, passwordHash, maxAttempts,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = repo.db.ExecContext(ctx,
			`UPDATE users SET reset_attempts = reset_attempts + 1
			WHERE email = $1 AND reset_code IS NOT NULL`,
			email,
		)
		if err != nil {
			return fmt.Errorf("tentative de réinitialisation : %w", err)
		}
		return ErrInvalidCode
	}
	if err != nil {
		return fmt.Errorf("réinitialisation du mot de passe : %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, id); err != nil {
		return fmt.Errorf("suppression des sessions : %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("réinitialisation du mot de passe : %w", err)
	}
	return nil
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
