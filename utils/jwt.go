package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/peekeah/book-store/config"
)

type JWTTokenBody struct {
	ID    uint
	Email string
	Name  string
}

type Token struct {
	UserId string  `json:"user_id"`
	Email  string  `json:"email"`
	Name   string  `json:"name"`
	Expiry float64 `json:"exp"`
	jwt.RegisteredClaims
}

func CreateJWTToken(user JWTTokenBody) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"expiry":  time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenStr, err := claims.SignedString([]byte(config.GetConfig().JWTSecretKey))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func VerifyJWTToken(tokenStr string) (uint, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte(config.GetConfig().JWTSecretKey), nil
	})
	if err != nil || !token.Valid {
		return 0, nil
	}

	if !token.Valid {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("")
	}

	id := uint(claims["user_id"].(float64))

	return id, nil
}
