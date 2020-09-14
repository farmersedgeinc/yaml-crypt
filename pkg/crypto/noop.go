package crypto

import "fmt"

type NoopProvider struct {
	Verbose bool `yaml:"verbose"`
}

func (p NoopProvider) Encrypt(plaintext string) ([]byte, error) {
	if p.Verbose {
		fmt.Printf("Encrypting %s", plaintext)
	}
	return []byte(plaintext), nil
}

func (p NoopProvider) Decrypt(ciphertext []byte) (string, error) {
	if p.Verbose {
		fmt.Printf("Decrypting %s", string(ciphertext))
	}
	return string(ciphertext), nil
}
