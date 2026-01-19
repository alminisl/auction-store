package password

import (
	"golang.org/x/crypto/bcrypt"
)

const cost = 12

// Hash generates a bcrypt hash from a password
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Verify compares a password with a hash
func Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
