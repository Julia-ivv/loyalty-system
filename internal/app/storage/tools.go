package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func hash(value, salt string) string {
	var s = append([]byte(value), []byte(salt)...)
	hash := sha256.Sum256(s)
	hashString := hex.EncodeToString(hash[:])

	return hashString
}

const LengthSalt = 4

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateRandomString(length int) (string, error) {
	b, err := GenerateRandomBytes(length)
	return base64.RawURLEncoding.EncodeToString(b), err
}
