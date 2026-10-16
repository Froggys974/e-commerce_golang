package user

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/Froggys974/e-commerce_golang/internal/ref"
)

type repository interface {
	List(ctx context.Context) ([]User, error)
	FindByID(ctx context.Context, id int64) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, id int64) error
	Confirm(ctx context.Context, id int64) error
	SetConfirmCode(ctx context.Context, email, code string, expiresAt time.Time) error
	ConfirmByCode(ctx context.Context, email, code string, maxAttempts int) error
	SetResetCode(ctx context.Context, email, code string, expiresAt time.Time) error
	ResetPassword(ctx context.Context, email, code, passwordHash string, maxAttempts int) error
}

type hasher func(password string) (string, error)

type checker func(hash, password string) error

const (
	minPasswordLength = 8
	maxPasswordLength = 72 // limite bcrypt
	codeLength        = 8
	codeTTL           = 15 * time.Minute
	maxCodeAttempts   = 5
)

type Service struct {
	repo  repository
	hash  hasher
	check checker
}

func NewService(repo repository, hashPassword hasher, checkPassword checker) *Service {
	return &Service{repo: repo, hash: hashPassword, check: checkPassword}
}

func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetUser(ctx context.Context, id int64) (User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Register(ctx context.Context, email, password string) (string, error) {
	user, err := s.create(ctx, email, password, RoleUser)
	if err != nil {
		return "", err
	}
	return s.ResendConfirmCode(ctx, user.Email)
}

// creation admin sans code confirmation via Confirm
func (s *Service) CreateUser(ctx context.Context, email, password string, role Role) (User, error) {
	return s.create(ctx, email, password, role)
}

func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	user, err := s.create(ctx, email, password, RoleAdmin)
	var emailTakenErr EmailTakenError
	if errors.As(err, &emailTakenErr) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.repo.Confirm(ctx, user.ID)
}

func (s *Service) create(ctx context.Context, email, password string, role Role) (User, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return User{}, err
	}
	if err := validateRole(role); err != nil {
		return User{}, err
	}
	if err := validatePassword(password); err != nil {
		return User{}, err
	}

	hash, err := s.hash(password)
	if err != nil {
		return User{}, err
	}

	user := User{
		Email:        email,
		PasswordHash: hash,
		Role:         role,
	}
	if err := s.repo.Create(ctx, &user); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, id int64, email string, role Role) (User, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return User{}, err
	}
	if err := validateRole(role); err != nil {
		return User{}, err
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return User{}, err
	}

	user.Email = email
	user.Role = role
	if err := s.repo.Update(ctx, user); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Service) DeleteUser(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Confirm(ctx context.Context, id int64) error {
	return s.repo.Confirm(ctx, id)
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (User, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if errors.Is(err, ErrEmailNotFound) {
		return User{}, ErrInvalidCredentials
	}
	if err != nil {
		return User{}, err
	}

	if err := s.check(user.PasswordHash, password); err != nil {
		return User{}, ErrInvalidCredentials
	}
	if !user.Confirmed {
		return User{}, ErrNotConfirmed
	}
	return user, nil
}

func (s *Service) ResendConfirmCode(ctx context.Context, email string) (string, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return "", err
	}

	code := ref.Code(codeLength)
	if err := s.repo.SetConfirmCode(ctx, email, ref.Hash(code), time.Now().Add(codeTTL)); err != nil {
		return "", err
	}
	return code, nil
}

func (s *Service) ConfirmByCode(ctx context.Context, email, code string) error {
	email, err := normalizeEmail(email)
	if err != nil {
		return err
	}

	code, err = normalizeCode(code)
	if err != nil {
		return err
	}
	return s.repo.ConfirmByCode(ctx, email, ref.Hash(code), maxCodeAttempts)
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return "", err
	}

	code := ref.Code(codeLength)
	if err := s.repo.SetResetCode(ctx, email, ref.Hash(code), time.Now().Add(codeTTL)); err != nil {
		return "", err
	}
	return code, nil
}

func (s *Service) ResetPassword(ctx context.Context, email, code, password string) error {
	email, err := normalizeEmail(email)
	if err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}

	code, err = normalizeCode(code)
	if err != nil {
		return err
	}

	hash, err := s.hash(password)
	if err != nil {
		return err
	}
	return s.repo.ResetPassword(ctx, email, ref.Hash(code), hash, maxCodeAttempts)
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", ValidationError{Field: "email", Message: "format invalide"}
	}
	return email, nil
}

func normalizeCode(code string) (string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != codeLength {
		return "", ErrInvalidCode
	}
	return code, nil
}

func validateRole(role Role) error {
	if !role.Valid() {
		return ValidationError{Field: "role", Message: "doit être user ou admin"}
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return ValidationError{Field: "password", Message: "doit être entre 8 et 72 octets"}
	}
	return nil
}
