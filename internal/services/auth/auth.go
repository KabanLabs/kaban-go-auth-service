package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
	"github.com/VACdotCS/kaban-go-auth-service/internal/lib/jwt"
	rsa_store "github.com/VACdotCS/kaban-go-auth-service/internal/lib/rsa-store"
	"github.com/VACdotCS/kaban-go-auth-service/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPublicKeyNotFound  = errors.New("public key not found")
)

type Auth struct {
	log                  *slog.Logger
	userSaver            UserSaver
	userProvider         UserProvider
	appProvider          AppProvider
	refreshTokenSaver    RefreshTokenSaver
	refreshTokenProvider RefreshTokenProvider
	accessTokenTTL       time.Duration
	refreshTokenTTL      time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid string, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

type RefreshTokenSaver interface {
	SaveRefreshToken(ctx context.Context, refreshToken string, uid string, expiresAt time.Time, appId int) (models.RefreshToken, error)
}

type RefreshTokenProvider interface {
	RefreshToken(ctx context.Context, refreshToken string) (models.RefreshToken, error)
}

// New returns new instance of the Auth Service
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	refreshTokenSaver RefreshTokenSaver,
	refreshTokenProvider RefreshTokenProvider,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:                  log,
		userSaver:            userSaver,
		userProvider:         userProvider,
		appProvider:          appProvider,
		refreshTokenProvider: refreshTokenProvider,
		refreshTokenSaver:    refreshTokenSaver,
		accessTokenTTL:       accessTokenTTL,
		refreshTokenTTL:      refreshTokenTTL,
	}
}

func (s *Auth) Login(ctx context.Context,
	email string,
	password string,
	appID int,
) (tokens TokenData, err error) {
	const op = "Auth.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login user")

	user, err := s.userProvider.User(ctx, email)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return TokenData{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		log.Error("failed to get user", "error", err)

		return TokenData{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid credentials", "error", err)

		return TokenData{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := s.appProvider.App(ctx, appID)

	if err != nil {
		return TokenData{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully logged in")

	accessToken, err := jwt.NewAccessToken(user, app, s.accessTokenTTL)

	if err != nil {
		log.Error("failed to generate access token", err)

		return TokenData{}, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := jwt.NewRefreshToken(user, app, s.refreshTokenTTL)

	if err != nil {
		log.Error("failed to generate refresh token", err)

		return TokenData{}, fmt.Errorf("%s: %w", op, err)
	}

	return TokenData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Auth) RegisterNewUser(ctx context.Context,
	email string,
	password string,
) (string, error) {
	const op = "Auth.RegisterNewUser"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to register user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error("failed to generate password", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	userID, err := s.userSaver.SaveUser(ctx, email, passHash)

	if err != nil {
		log.Error("failed to save user", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (s *Auth) ValidateAccessToken(ctx context.Context,
	accessToken string,
) (isValid bool, err error) {
	panic("implement me")
}

func (s *Auth) GetJWK(ctx context.Context,
	kid string,
) (key *models.JWK, err error) {
	key, err = rsa_store.GetJWKByKid(kid)

	if err != nil {
		return nil, err
	}

	return key, nil
}

func (s *Auth) RegenerateRefreshToken(ctx context.Context,
	refreshToken string,
) (accessToken string, err error) {
	panic("implement me")
}

type TokenData struct {
	AccessToken  string
	RefreshToken string
}
