package jwt_actions

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v5/pgxpool"
)

var jwtKey = []byte("some_secret_key")

type Claims struct {
	Username string `json:"username"`
	UserID   string `json:"user_id"`
	jwt.StandardClaims
}

func GenerateJWT(db *pgxpool.Pool, username string, userID string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		UserID:   userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, errors.New("invalid token signature")
		}
		return nil, errors.New("could not parse token")
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
