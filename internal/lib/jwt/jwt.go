package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
	"github.com/VACdotCS/kaban-go-auth-service/internal/lib/rsa-store"
	"github.com/golang-jwt/jwt/v5"
)

// NewAccessToken Generates new access token
func NewAccessToken(user models.User, app models.App, duration time.Duration, privateKey *rsa.PrivateKey) (string, error) {
	token := jwt.New(jwt.SigningMethodRS512)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["iat"] = time.Now().Unix()
	claims["app_id"] = app.ID
	claims["kid"] = rsa_store.GetLastKeyId()

	tokenString, err := token.SignedString(privateKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// NewRefreshToken Just alias for refresh token creation, do the same as NewAccessToken but with another duration
func NewRefreshToken(user models.User, app models.App, duration time.Duration, privateKey *rsa.PrivateKey) (string, error) {
	return NewAccessToken(user, app, duration, privateKey)
}

func CheckToken(tokenString string) (bool, error) {
	decodedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Claims.(jwt.MapClaims)["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		jwk, err := rsa_store.GetJWKByKid(kid)
		if err != nil {
			return nil, fmt.Errorf("jwk not found for kid=%s: %w", kid, err)
		}

		pubKey, err := jwk.ToPublicKey()
		if err != nil {
			return nil, fmt.Errorf("invalid jwk to rsa.PublicKey: %w", err)
		}

		return pubKey, nil
	})

	if err != nil {
		return false, err
	}

	if !decodedToken.Valid {
		return false, fmt.Errorf("token is invalid")
	}

	claims, ok := decodedToken.Claims.(jwt.MapClaims)

	if !ok {
		return false, fmt.Errorf("invalid claims type")
	}

	// Проверка exp
	if expRaw, ok := claims["exp"]; ok {
		switch exp := expRaw.(type) {
		case float64: // стандартный json.Unmarshal
			if int64(exp) < time.Now().Unix() {
				return false, fmt.Errorf("token expired")
			}
		case int64:
			if exp < time.Now().Unix() {
				return false, fmt.Errorf("token expired")
			}
		default:
			return false, fmt.Errorf("invalid exp format")
		}
	}

	return true, nil
}
