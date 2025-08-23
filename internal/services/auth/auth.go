package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
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
	App(ctx context.Context, appID string) (models.App, error)
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
	panic("implement me")
}

func (s *Auth) RegisterNewUser(ctx context.Context,
	email string,
	password string,
) (userID string, err error) {
	panic("implement me")
}

func (s *Auth) ValidateAccessToken(ctx context.Context,
	accessToken string,
) (isValid bool, err error) {
	//_ := jwt.ValidateTokenWithClaims(accessToken)
	panic("implement me")
}

func (s *Auth) RegenerateAccessToken(ctx context.Context,
	refreshToken string,
) (accessToken string, err error) {
	panic("implement me")
}

type TokenData struct {
	AccessToken  string
	RefreshToken string
}
