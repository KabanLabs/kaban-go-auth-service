package auth

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
	"github.com/VACdotCS/kaban-go-auth-service/internal/lib/hash"
	rsa_store "github.com/VACdotCS/kaban-go-auth-service/internal/lib/rsa-store"
	"github.com/VACdotCS/kaban-go-auth-service/internal/storage"
)

type mockDeps struct {
	user     models.User
	userErr  error
	saveErr  error
	savedUid string
}

func (m *mockDeps) User(ctx context.Context, email string) (models.User, error) {
	return m.user, m.userErr
}

func (m *mockDeps) UserById(ctx context.Context, uid string) (models.User, error) {
	return m.user, m.userErr
}

func (m *mockDeps) SaveUser(ctx context.Context, email string, passHash []byte) (string, error) {
	return m.savedUid, m.saveErr
}

func (m *mockDeps) App(ctx context.Context, appID int) (models.App, error) {
	return models.App{ID: appID, Name: "TestApp", Scopes: "all"}, nil
}

func (m *mockDeps) SaveRefreshToken(ctx context.Context, refreshToken string, uid string, expiresAt time.Time, appId int) (*models.RefreshToken, error) {
	return &models.RefreshToken{}, nil
}

func (m *mockDeps) RefreshToken(ctx context.Context, refreshToken string) (models.RefreshToken, error) {
	return models.RefreshToken{}, nil
}

func (m *mockDeps) RotateRefreshToken(ctx context.Context, refreshToken string) error {
	return nil
}

func TestAuth_RegisterNewUser(t *testing.T) {
	log := slog.Default()
	deps := &mockDeps{savedUid: "user-123"}
	svc := New(log, deps, deps, deps, deps, deps, deps, time.Minute, time.Hour)

	t.Run("Success", func(t *testing.T) {
		uid, err := svc.RegisterNewUser(context.Background(), "test@test.com", "strongpassword123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if uid != "user-123" {
			t.Errorf("expected uid user-123, got %s", uid)
		}
	})

	t.Run("Email Already Exists", func(t *testing.T) {
		deps.saveErr = storage.ErrUserExists
		_, err := svc.RegisterNewUser(context.Background(), "test@test.com", "strongpassword123")
		if !errors.Is(err, ErrEmailAlreadyExists) {
			t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
		}
	})
}

func TestAuth_Login(t *testing.T) {
	_, _ = rsa_store.LoadOrGenerateKeys(2048, 1*time.Hour)

	log := slog.Default()

	passHash, _ := hash.GenerateFromPassword("my_secret", nil)

	deps := &mockDeps{
		user: models.User{ID: "user-123", PassHash: []byte(passHash)},
	}

	svc := New(log, deps, deps, deps, deps, deps, deps, time.Minute, time.Hour)

	t.Run("Valid Credentials", func(t *testing.T) {
		tokens, err := svc.Login(context.Background(), "test@test.com", "my_secret", 1)
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if tokens.AccessToken == "" || tokens.RefreshToken == "" {
			t.Error("expected tokens to be generated")
		}
	})

	t.Run("Invalid Password", func(t *testing.T) {
		_, err := svc.Login(context.Background(), "test@test.com", "wrong", 1)
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		deps.userErr = storage.ErrUserNotFound
		_, err := svc.Login(context.Background(), "notfound@test.com", "any", 1)

		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}
