package memory

import (
	"sync"

	"github.com/farmersedgeinc/yaml-crypt/pkg/cache/common"
)

type memoryCache struct {
	c map[string]string
	m sync.RWMutex
}

func Setup() (*memoryCache, error) {
	return &memoryCache{
		c: map[string]string{},
	}, nil
}

func (c *memoryCache) Close() error {
	return nil
}

func (c *memoryCache) Add(plaintext string, ciphertext []byte) error {
	c.m.Lock()
	defer c.m.Unlock()
	c.c[string(common.PlaintextToKey(plaintext))] = string(ciphertext)
	c.c[string(common.CiphertextToKey(ciphertext))] = plaintext
	return nil
}

func (c *memoryCache) Encrypt(plaintext string, potentialCiphertext []byte) ([]byte, bool, error) {
	c.m.RLock()
	defer c.m.RUnlock()
	// if the potentialCiphertext is in the cache, and has a plaintext equal to the plaintext being encrypted, that's the ciphertext!
	if len(potentialCiphertext) > 0 {
		potentialCiphertextPlaintext, ok := c.c[string(common.CiphertextToKey(potentialCiphertext))]
		if ok && string(potentialCiphertextPlaintext) == plaintext {
			return potentialCiphertext, ok, nil
		}
	}
	// potentialCiphertext wasn't it, so return an arbitrary ciphertext that encrypts the given plaintext.
	ciphertext, ok := c.c[string(common.PlaintextToKey(plaintext))]
	return []byte(ciphertext), ok, nil
}

func (c *memoryCache) Decrypt(ciphertext []byte) (string, bool, error) {
	c.m.RLock()
	defer c.m.RUnlock()
	plaintext, ok := c.c[string(common.CiphertextToKey(ciphertext))]
	return plaintext, ok, nil
}
