package grpcapp

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
	"github.com/VACdotCS/kaban-go-auth-service/internal/services/auth"
)

type mockAuth struct{}

func (m *mockAuth) Login(ctx context.Context, email string, password string, appID int) (auth.TokenData, error) {
	return auth.TokenData{}, nil
}
func (m *mockAuth) RegisterNewUser(ctx context.Context, email string, password string) (string, error) {
	return "", nil
}
func (m *mockAuth) GetJWK(kid string) (*models.JWK, error) {
	return nil, nil
}
func (m *mockAuth) ValidateAccessToken(accessToken string) (bool, error) {
	return false, nil
}
func (m *mockAuth) RegenerateAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	return "", "", nil
}

func TestGRPCApp_GracefulShutdown(t *testing.T) {
	log := slog.Default()

	// Create app on a random port (port 0 lets the OS pick)
	app := New(log, 0, &mockAuth{})

	go func() {
		_ = app.Run()
	}()

	// Wait for server to initialize
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// The stop should finish instantly and not block forever
	app.Stop(ctx)
}
