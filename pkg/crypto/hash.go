package crypto

import (
	"crypto/rand"
	"crypto/sha256"
)

func Salt() ([]byte, error) {
	result := make([]byte, 16)
	_, err := rand.Read(result)
	return result, err
}

func Hash(salt []byte, data string) []byte {
	saltCopy := make([]byte, len(salt))
	copy(saltCopy, salt)
	result := sha256.Sum256(append(saltCopy, []byte(data)...))
	return result[:16]
}
