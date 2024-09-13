package password_hasher

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword получает на вход пароль и возвращает его хэш в виде строки или ошибку
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
