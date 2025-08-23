package jwt

import (
	"time"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
)

// NewAccessToken Generates new access token
func NewAccessToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodRS512)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["iat"] = time.Now().Unix()
	claims["app_id"] = app.ID

	// TODO: pass RSA private key here
	tokenString, err := token.SignedString("")

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// NewRefreshToken Just alias for refresh token creation, do the same as NewAccessToken but with another duration
func NewRefreshToken(user models.User, app models.App, duration time.Duration) (string, error) {
	return NewAccessToken(user, app, duration)
}
