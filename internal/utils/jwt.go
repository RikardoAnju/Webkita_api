// utils/jwt.go
package utils

import (
	"time"
	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte("your-secret-key") // Ganti dengan secret key yang aman

type Claims struct {
	UserID interface{} `json:"user_id"`
	Email  string      `json:"email"`
	jwt.StandardClaims
}

func GenerateJWT(userID interface{}, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}