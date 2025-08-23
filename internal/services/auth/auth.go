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
	ErrEmailAlreadyExists = errors.New("email already exists")
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
	UserById(ctx context.Context, uid string) (models.User, error)
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
			return TokenData{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
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
		log.Error("failed to get app", "error", err)
		return TokenData{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully logged in")

	privateKey := rsa_store.GetPrivateKey()

	accessToken, err := jwt.NewAccessToken(user, app, s.accessTokenTTL, &privateKey)

	if err != nil {
		log.Error("failed to generate access token", err)

		return TokenData{}, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := jwt.NewRefreshToken(user, app, s.refreshTokenTTL, &privateKey)

	if err != nil {
		log.Error("failed to generate refresh token", err)

		return TokenData{}, fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.refreshTokenSaver.SaveRefreshToken(ctx, refreshToken, user.ID, time.Now().Add(s.refreshTokenTTL), appID)

	if err != nil {
		log.Error("failed to save refresh token", err)
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

		if errors.Is(err, storage.ErrUserExists) {
			return "", fmt.Errorf("%s: %w", op, ErrEmailAlreadyExists)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (s *Auth) ValidateAccessToken(accessToken string) (isValid bool, err error) {
	const op = "Auth.ValidateAccessToken"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to validate access token")

	valid, err := jwt.CheckToken(accessToken)

	if err != nil {
		log.Error("failed to validate access token", err)
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return valid, nil
}

func (s *Auth) GetJWK(kid string) (key *models.JWK, err error) {
	const op = "Auth.GetJWK"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info(fmt.Sprintf("Getting public key with kid = %s", kid))

	key, err = rsa_store.GetJWKByKid(kid)

	if err != nil {
		log.Error("failed to get jwk", ErrPublicKeyNotFound)
		return nil, fmt.Errorf("%s: %w", op, ErrPublicKeyNotFound)
	}

	return key, nil
}

func (s *Auth) RegenerateRefreshToken(
	ctx context.Context,
	refreshToken string,
) (string, string, error) {
	const op = "Auth.RegenerateRefreshToken"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to regenerate tokens")

	valid, err := jwt.CheckToken(refreshToken)

	if err != nil || !valid {
		log.Error("invalid refresh token", "error", err)
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	storedRefresh, err := s.refreshTokenProvider.RefreshToken(ctx, refreshToken)
	if err != nil {
		log.Error("refresh token not found", "error", err)
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	if time.Now().After(storedRefresh.ExpireAt) {
		log.Error("refresh token expired")
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	user, err := s.userProvider.UserById(ctx, storedRefresh.UserID)
	if err != nil {
		log.Error("failed to get user", "error", err)
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	app, err := s.appProvider.App(ctx, storedRefresh.AppID)
	if err != nil {
		log.Error("failed to get app", "error", err)
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	privateKey := rsa_store.GetPrivateKey()

	newAccessToken, err := jwt.NewAccessToken(user, app, s.accessTokenTTL, &privateKey)

	if err != nil {
		log.Error("failed to generate access token", "error", err)
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	newRefreshToken, err := jwt.NewRefreshToken(user, app, s.refreshTokenTTL, &privateKey)

	if err != nil {
		log.Error("failed to generate refresh token", "error", err)
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.refreshTokenSaver.SaveRefreshToken(
		ctx,
		newRefreshToken,
		user.ID,
		time.Now().Add(s.refreshTokenTTL),
		app.ID,
	)

	if err != nil {
		log.Error("failed to save refresh token", "error", err)
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return newAccessToken, newRefreshToken, nil
}

type TokenData struct {
	AccessToken  string
	RefreshToken string
}
