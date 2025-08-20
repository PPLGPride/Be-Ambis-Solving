package utils

import (
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(sub string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   sub,
		Issuer:    "be-ambis-solving",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(config.Cfg.JWTSecret))
}

func ParseJWT(tokenStr string) (*jwt.RegisteredClaims, error) {
	tkn, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.Cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tkn.Claims.(*jwt.RegisteredClaims); ok && tkn.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
