package token

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateTokens(userID int64) (string, string, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	})
	accessString, err := accessToken.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	refreshString, err := refreshToken.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	return accessString, refreshString, nil
}
