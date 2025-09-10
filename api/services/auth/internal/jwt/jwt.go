package jwt

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type AccessClaims struct {
	UserID uint   `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID uint, email string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(), // jti
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    "auth-service",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseAndValidateAccessToken(tok string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tok, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*AccessClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token claims")
}
