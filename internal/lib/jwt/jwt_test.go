package jwt

import (
	"strings"
	"testing"
	"time"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
	rsa_store "github.com/VACdotCS/kaban-go-auth-service/internal/lib/rsa-store"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	keys, err := rsa_store.LoadOrGenerateKeys(2048, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to init keys: %v", err)
	}
	privateKey := keys.PrivateKey

	user := models.User{ID: "user-123"}
	app := models.App{ID: 1}

	t.Run("Valid Token", func(t *testing.T) {
		token, err := NewAccessToken(user, app, 10*time.Minute, privateKey)
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		valid, err := CheckToken(token)
		if err != nil {
			t.Fatalf("token validation failed: %v", err)
		}
		if !valid {
			t.Error("expected token to be valid")
		}
	})

	t.Run("Expired Token", func(t *testing.T) {
		expiredToken, err := NewAccessToken(user, app, -1*time.Second, privateKey)
		if err != nil {
			t.Fatalf("failed to create expired token: %v", err)
		}

		valid, err := CheckToken(expiredToken)
		if err == nil {
			t.Error("expected error for expired token, got nil")
		}
		if valid {
			t.Error("expected expired token to be invalid")
		}
	})

	t.Run("Forged Signature", func(t *testing.T) {
		token, _ := NewAccessToken(user, app, 10*time.Minute, privateKey)

		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Fatalf("invalid token format")
		}

		sig := parts[2]
		if sig[0] == 'A' {
			sig = "B" + sig[1:]
		} else {
			sig = "A" + sig[1:]
		}

		forgedToken := parts[0] + "." + parts[1] + "." + sig

		valid, err := CheckToken(forgedToken)
		if err == nil {
			t.Error("expected error for forged signature, got nil")
		}
		if valid {
			t.Error("expected forged token to be invalid")
		}
	})
}
