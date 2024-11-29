package password

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashingPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password is empty")
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("HashingPassword: failed to generate password to bcrypt hash, msg:", err)
		return "", fmt.Errorf("server error")
	}
	return string(hashedPass), err
}

func VerifyHashPassword(password string, hashedPass string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password))
}
