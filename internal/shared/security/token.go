package security

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
)

func GenerateToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func BuildConfirmLink(baseURL, token string) string {
	return baseURL + "/confirm?token=" + token
}

func GenerateTicketToken() string {
	return uuid.NewString()
}