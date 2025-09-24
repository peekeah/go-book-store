package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserBody struct {
	Id    string
	Email string
	Name  string
}

var secretKey = []byte("secret-key")

func CreateJWTToken(user UserBody) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":    user.Id,
			"email": user.Email,
			"name":  user.Name,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})

	return token.SignedString(secretKey)
}

func VerifyJWTToken(token string) (string, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &jwt.MapClaims{}, func(toke *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}

	if !parsedToken.Valid {
		return "", errors.New("unauthorized")
	}

	claims, ok := parsedToken.Claims.(*jwt.MapClaims)

	if !ok {
		return "", errors.New("failed to parse token")
	}

	id := (*claims)["id"].(string)

	return id, err
}
