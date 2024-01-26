package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomString() string {
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(randomBytes)[:16]
}
