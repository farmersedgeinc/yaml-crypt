package cache

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache/disk"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache/memory"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
)

type Cache interface {
	Close() error
	Encrypt(plaintext string, potentialCiphertext []byte) ([]byte, bool, error)
	Decrypt(ciphertext []byte) (string, bool, error)
	Add(plaintext string, ciphertext []byte) error
}

func Setup(config config.Config, mem bool) (Cache, error) {
	if mem {
		return memory.Setup()
	} else {
		return disk.Setup(config)
	}
}
