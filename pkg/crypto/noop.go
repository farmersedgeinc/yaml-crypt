package crypto

import (
	"fmt"
	"time"
)

type NoopProvider struct {
	Verbose bool `yaml:"verbose"`
}

func (p NoopProvider) Encrypt(plaintext string, _ uint, _ time.Duration) ([]byte, error) {
	if p.Verbose {
		fmt.Printf("Encrypting %s", plaintext)
	}
	return []byte(plaintext), nil
}

func (p NoopProvider) Decrypt(ciphertext []byte, _ uint, _ time.Duration) (string, error) {
	if p.Verbose {
		fmt.Printf("Decrypting %s", string(ciphertext))
	}
	return string(ciphertext), nil
}
