package user

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Froggys974/e-commerce_golang/internal/ref"
)

type fakeRepository struct {
	users       map[int64]User
	codes       map[int64]string
	resetCodes  map[int64]string
	nextID      int64
	createErr   error
	updateCalls int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		users:      map[int64]User{},
		codes:      map[int64]string{},
		resetCodes: map[int64]string{},
		nextID:     1,
	}
}

func (f *fakeRepository) seed(u User) User {
	u.ID = f.nextID
	f.nextID++
	f.users[u.ID] = u
	return u
}

func (f *fakeRepository) List(ctx context.Context) ([]User, error) {
	users := make([]User, 0, len(f.users))
	for _, u := range f.users {
		users = append(users, u)
	}
	return users, nil
}

func (f *fakeRepository) FindByID(ctx context.Context, id int64) (User, error) {
	u, ok := f.users[id]
	if !ok {
		return User{}, NotFoundError{ID: id}
	}
	return u, nil
}

func (f *fakeRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return User{}, ErrEmailNotFound
}

func (f *fakeRepository) Create(ctx context.Context, u *User) error {
	if f.createErr != nil {
		return f.createErr
	}
	u.ID = f.nextID
	u.CreatedAt = time.Now()
	f.nextID++
	f.users[u.ID] = *u
	return nil
}

func (f *fakeRepository) Update(ctx context.Context, u User) error {
	if _, ok := f.users[u.ID]; !ok {
		return NotFoundError{ID: u.ID}
	}
	f.users[u.ID] = u
	f.updateCalls++
	return nil
}

func (f *fakeRepository) Delete(ctx context.Context, id int64) error {
	if _, ok := f.users[id]; !ok {
		return NotFoundError{ID: id}
	}
	delete(f.users, id)
	return nil
}

func (f *fakeRepository) Confirm(ctx context.Context, id int64) error {
	u, ok := f.users[id]
	if !ok {
		return NotFoundError{ID: id}
	}
	u.Confirmed = true
	f.users[id] = u
	delete(f.codes, id)
	return nil
}

func (f *fakeRepository) SetConfirmCode(ctx context.Context, email, code string, expiresAt time.Time) error {
	u, err := f.FindByEmail(ctx, email)
	if err != nil || u.Confirmed {
		return ErrEmailNotFound
	}
	f.codes[u.ID] = code
	return nil
}

func (f *fakeRepository) ConfirmByCode(ctx context.Context, email, code string, maxAttempts int) error {
	u, err := f.FindByEmail(ctx, email)
	if err != nil || u.Confirmed || f.codes[u.ID] != code {
		return ErrInvalidCode
	}
	return f.Confirm(ctx, u.ID)
}

func (f *fakeRepository) SetResetCode(ctx context.Context, email, code string, expiresAt time.Time) error {
	u, err := f.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	f.resetCodes[u.ID] = code
	return nil
}

func (f *fakeRepository) ResetPassword(ctx context.Context, email, code, passwordHash string, maxAttempts int) error {
	u, err := f.FindByEmail(ctx, email)
	if err != nil || f.resetCodes[u.ID] != code {
		return ErrInvalidCode
	}
	u.PasswordHash = passwordHash
	f.users[u.ID] = u
	delete(f.resetCodes, u.ID)
	return nil
}

func fakeHash(password string) (string, error) {
	return "hashed:" + password, nil
}

func fakeCheck(hash, password string) error {
	if hash != "hashed:"+password {
		return errors.New("mauvais mot de passe")
	}
	return nil
}

func newTestService(repo *fakeRepository) *Service {
	return NewService(repo, fakeHash, fakeCheck)
}

func TestCreateUser(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)

	u, err := svc.CreateUser(context.Background(), "  Florent@Test.COM ", "motdepasse", RoleAdmin)
	if err != nil {
		t.Fatalf("erreur : %v", err)
	}

	if u.ID == 0 {
		t.Error("l'id aurait dû être rempli par le repo")
	}
	if u.Email != "florent@test.com" {
		t.Errorf("Email = %q; attendu %q", u.Email, "florent@test.com")
	}
	if u.PasswordHash != "hashed:motdepasse" {
		t.Errorf("PasswordHash = %q; attendu le mot de passe hashé", u.PasswordHash)
	}
	if u.Confirmed {
		t.Error("un nouvel utilisateur ne devrait pas être confirmé")
	}
	if _, ok := repo.users[u.ID]; !ok {
		t.Error("l'utilisateur n'a pas été enregistré dans le repo")
	}
	if _, ok := repo.codes[u.ID]; ok {
		t.Error("un utilisateur créé par l'admin ne devrait pas avoir de code")
	}
}

func TestCreateUserValidation(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		role      Role
		wantField string
	}{
		{"email invalide", "pas-un-email", "motdepasse", RoleUser, "email"},
		{"email vide", "", "motdepasse", RoleUser, "email"},
		{"email avec nom affiché", "Florent <florent@test.com>", "motdepasse", RoleUser, "email"},
		{"rôle inconnu", "florent@test.com", "motdepasse", Role("superadmin"), "role"},
		{"rôle vide", "florent@test.com", "motdepasse", Role(""), "role"},
		{"mot de passe trop court", "florent@test.com", "court", RoleUser, "password"},
		{"mot de passe trop long", "florent@test.com", strings.Repeat("a", 73), RoleUser, "password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			svc := newTestService(repo)

			_, err := svc.CreateUser(context.Background(), tt.email, tt.password, tt.role)

			var valErr ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("erreur = %v; attendu une ValidationError", err)
			}
			if valErr.Field != tt.wantField {
				t.Errorf("Field = %q; attendu %q", valErr.Field, tt.wantField)
			}
			if len(repo.users) != 0 {
				t.Error("aucun utilisateur ne devrait être enregistré")
			}
		})
	}
}

func TestCreateUserHashError(t *testing.T) {
	repo := newFakeRepository()
	errHash := errors.New("hash impossible")
	svc := NewService(repo, func(string) (string, error) {
		return "", errHash
	}, fakeCheck)

	_, err := svc.CreateUser(context.Background(), "florent@test.com", "motdepasse", RoleUser)

	if !errors.Is(err, errHash) {
		t.Errorf("erreur = %v; attendu %v", err, errHash)
	}
	if len(repo.users) != 0 {
		t.Error("aucun utilisateur ne devrait être enregistré")
	}
}

func TestCreateUserRepositoryError(t *testing.T) {
	repo := newFakeRepository()
	repo.createErr = EmailTakenError{Email: "florent@test.com"}
	svc := newTestService(repo)

	_, err := svc.CreateUser(context.Background(), "florent@test.com", "motdepasse", RoleUser)

	var takenErr EmailTakenError
	if !errors.As(err, &takenErr) {
		t.Errorf("erreur = %v; attendu une EmailTakenError", err)
	}
}

func TestRegister(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)

	code, err := svc.Register(context.Background(), "Florent@Test.com", "motdepasse")
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}

	if len(code) != codeLength {
		t.Errorf("len(code) = %d; attendu %d", len(code), codeLength)
	}
	if len(repo.users) != 1 {
		t.Fatalf("len(users) = %d; attendu 1", len(repo.users))
	}
	for id, u := range repo.users {
		if u.Role != RoleUser {
			t.Errorf("Role = %q; attendu %q", u.Role, RoleUser)
		}
		if repo.codes[id] != ref.Hash(code) {
			t.Errorf("code stocké = %q; attendu %q", repo.codes[id], ref.Hash(code))
		}
	}
}

func TestRegisterError(t *testing.T) {
	repo := newFakeRepository()
	repo.createErr = EmailTakenError{Email: "florent@test.com"}
	svc := newTestService(repo)

	code, err := svc.Register(context.Background(), "florent@test.com", "motdepasse")

	var takenErr EmailTakenError
	if !errors.As(err, &takenErr) {
		t.Errorf("erreur = %v; attendu une EmailTakenError", err)
	}
	if code != "" {
		t.Errorf("code = %q; attendu vide en cas d'erreur", code)
	}
}

func TestGetUser(t *testing.T) {
	repo := newFakeRepository()
	existing := repo.seed(User{Email: "florent@test.com", Role: RoleUser})
	svc := newTestService(repo)

	u, err := svc.GetUser(context.Background(), existing.ID)
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if u.Email != "florent@test.com" {
		t.Errorf("Email = %q; attendu %q", u.Email, "florent@test.com")
	}

	_, err = svc.GetUser(context.Background(), 999)
	var notFound NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("erreur = %v; attendu une NotFoundError", err)
	}
}

func TestListUsers(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(User{Email: "florent@test.com", Role: RoleUser})
	repo.seed(User{Email: "bob@test.com", Role: RoleAdmin})
	svc := newTestService(repo)

	users, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if len(users) != 2 {
		t.Errorf("len(users) = %d; attendu 2", len(users))
	}
}

func TestUpdateUser(t *testing.T) {
	repo := newFakeRepository()
	existing := repo.seed(User{Email: "florent@test.com", PasswordHash: "hashed:secret", Role: RoleUser})
	svc := newTestService(repo)

	u, err := svc.UpdateUser(context.Background(), existing.ID, "Florent.New@Test.com", RoleAdmin)
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}

	if u.Email != "florent.new@test.com" {
		t.Errorf("Email = %q; attendu %q", u.Email, "florent.new@test.com")
	}
	if u.Role != RoleAdmin {
		t.Errorf("Role = %q; attendu %q", u.Role, RoleAdmin)
	}
	if repo.users[existing.ID].PasswordHash != "hashed:secret" {
		t.Error("la mise à jour ne devrait pas toucher au mot de passe")
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	svc := newTestService(newFakeRepository())

	_, err := svc.UpdateUser(context.Background(), 42, "florent@test.com", RoleUser)

	var notFound NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("erreur = %v; attendu une NotFoundError", err)
	}
	if notFound.ID != 42 {
		t.Errorf("ID = %d; attendu 42", notFound.ID)
	}
}

func TestUpdateUserValidation(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		role      Role
		wantField string
	}{
		{"email invalide", "pas-un-email", RoleUser, "email"},
		{"rôle inconnu", "florent@test.com", Role("superadmin"), "role"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			existing := repo.seed(User{Email: "florent@test.com", Role: RoleUser})
			svc := newTestService(repo)

			_, err := svc.UpdateUser(context.Background(), existing.ID, tt.email, tt.role)

			var valErr ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("erreur = %v; attendu une ValidationError", err)
			}
			if valErr.Field != tt.wantField {
				t.Errorf("Field = %q; attendu %q", valErr.Field, tt.wantField)
			}
			if repo.updateCalls != 0 {
				t.Error("le repo ne devrait pas être appelé si la validation échoue")
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	repo := newFakeRepository()
	existing := repo.seed(User{Email: "florent@test.com", Role: RoleUser})
	svc := newTestService(repo)

	if err := svc.DeleteUser(context.Background(), existing.ID); err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if _, ok := repo.users[existing.ID]; ok {
		t.Error("l'utilisateur aurait dû être supprimé")
	}

	err := svc.DeleteUser(context.Background(), existing.ID)
	var notFound NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("erreur = %v; attendu une NotFoundError", err)
	}
}

func TestConfirm(t *testing.T) {
	repo := newFakeRepository()
	existing := repo.seed(User{Email: "florent@test.com", Role: RoleUser})
	repo.codes[existing.ID] = "ABCD2345"
	svc := newTestService(repo)

	if err := svc.Confirm(context.Background(), existing.ID); err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}

	if !repo.users[existing.ID].Confirmed {
		t.Error("l'utilisateur devrait être confirmé")
	}
	if _, ok := repo.codes[existing.ID]; ok {
		t.Error("le code de confirmation devrait être effacé")
	}
}

func TestConfirmNotFound(t *testing.T) {
	svc := newTestService(newFakeRepository())

	err := svc.Confirm(context.Background(), 42)

	var notFound NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("erreur = %v; attendu une NotFoundError", err)
	}
}

func TestConfirmByCode(t *testing.T) {
	repo := newFakeRepository()
	existing := repo.seed(User{Email: "florent@test.com", Role: RoleUser})
	repo.codes[existing.ID] = ref.Hash("ABCD2345")
	svc := newTestService(repo)

	err := svc.ConfirmByCode(context.Background(), " Florent@Test.com ", " abcd2345 ")
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if !repo.users[existing.ID].Confirmed {
		t.Error("l'utilisateur devrait être confirmé")
	}

	err = svc.ConfirmByCode(context.Background(), "florent@test.com", "ABCD2345")
	if !errors.Is(err, ErrInvalidCode) {
		t.Errorf("erreur = %v; attendu ErrInvalidCode", err)
	}
}

func TestConfirmByCodeInvalid(t *testing.T) {
	tests := []struct {
		name  string
		email string
		code  string
	}{
		{"mauvais code", "florent@test.com", "ZZZZ2345"},
		{"code vide", "florent@test.com", ""},
		{"code trop court", "florent@test.com", "ABCD"},
		{"autre email", "bob@test.com", "ABCD2345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			existing := repo.seed(User{Email: "florent@test.com", Role: RoleUser})
			repo.codes[existing.ID] = ref.Hash("ABCD2345")
			svc := newTestService(repo)

			err := svc.ConfirmByCode(context.Background(), tt.email, tt.code)

			if !errors.Is(err, ErrInvalidCode) {
				t.Errorf("erreur = %v; attendu ErrInvalidCode", err)
			}
			if repo.users[existing.ID].Confirmed {
				t.Error("l'utilisateur ne devrait pas être confirmé")
			}
		})
	}
}

func TestConfirmByCodeBadEmail(t *testing.T) {
	svc := newTestService(newFakeRepository())

	err := svc.ConfirmByCode(context.Background(), "pas-un-email", "ABCD2345")

	var valErr ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("erreur = %v; attendu une ValidationError", err)
	}
}

func TestAuthenticate(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(User{Email: "florent@test.com", PasswordHash: "hashed:motdepasse", Confirmed: true})
	repo.seed(User{Email: "bob@test.com", PasswordHash: "hashed:motdepasse"})
	svc := newTestService(repo)

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{"ok", "florent@test.com", "motdepasse", nil},
		{"mauvais mot de passe", "florent@test.com", "mauvais!!", ErrInvalidCredentials},
		{"email inconnu", "inconnu@test.com", "motdepasse", ErrInvalidCredentials},
		{"compte non confirmé", "bob@test.com", "motdepasse", ErrNotConfirmed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Authenticate(context.Background(), tt.email, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("erreur = %v; attendu %v", err, tt.wantErr)
			}
		})
	}
}

func TestResetPassword(t *testing.T) {
	repo := newFakeRepository()
	existing := repo.seed(User{Email: "florent@test.com", PasswordHash: "hashed:ancien"})
	svc := newTestService(repo)

	code, err := svc.RequestPasswordReset(context.Background(), "florent@test.com")
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}

	err = svc.ResetPassword(context.Background(), "florent@test.com", "ZZZZZZZZ", "nouveaumdp")
	if !errors.Is(err, ErrInvalidCode) {
		t.Errorf("mauvais code : erreur = %v; attendu ErrInvalidCode", err)
	}

	if err := svc.ResetPassword(context.Background(), "florent@test.com", code, "nouveaumdp"); err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if repo.users[existing.ID].PasswordHash != "hashed:nouveaumdp" {
		t.Error("le mot de passe aurait dû être modifié")
	}
}
