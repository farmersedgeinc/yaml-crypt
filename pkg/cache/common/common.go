package common

import "crypto/sha256"

const (
	// Length to hash plaintext and ciphertext keys.
	hashLength = 16
	// Prefix for keys containing a hashed plaintext, used to look up ciphertext.
	plaintextKeyPrefix = 'p'
	// Prefix for keys containing a hashed ciphertext, used to look up plaintext.
	ciphertextKeyPrefix = 'c'
)

// Convert a plaintext to the key used to lookup its ciphertext.
func PlaintextToKey(data string) []byte {
	key := make([]byte, 1, hashLength+1)
	key[0] = plaintextKeyPrefix
	key = append(key, hash([]byte(data))...)
	return key
}

// Hash some bytes, truncating the length to the hashLength constant.
func hash(data []byte) []byte {
	result := sha256.Sum256(data)
	return result[:hashLength]
}

// Convert a ciphertext to the key used to lookup its plaintext.
func CiphertextToKey(data []byte) []byte {
	key := make([]byte, 1, hashLength+1)
	key[0] = ciphertextKeyPrefix
	key = append(key, hash(data)...)
	return key
}
