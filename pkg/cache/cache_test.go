package cache

import (
	"fmt"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/fixtures"
	"testing"
)

func TestCache(t *testing.T) {
	// make the cache a lot smaller to make it quicker to test LRU behavior
	YoungCacheSize = 1024 * 200
	// check out an arbitrary repo in order to provide a directory and config for the cache
	repos, err := fixtures.Repos()
	if err != nil {
		t.Fatal(err)
	}
	repo := repos[0]
	err = repo.Setup()
	defer repo.Destroy()
	if err != nil {
		t.Fatal(err)
	}
	config, err := config.LoadConfig(".")
	if err != nil {
		t.Fatal(err)
	}
	cache, err := Setup(config)
	if err != nil {
		t.Fatal(err)
	}
	encryptNonExistent(t, &cache)
	decryptNonExistent(t, &cache)
	cache.Close()
	for round := 0; round < 20; round++ {
		cache, err := Setup(config)
		if err != nil {
			t.Fatal(err)
		}
		for item := 0; item < 100; item++ {
			err := cache.Add(
				plaintext(round, item),
				ciphertext(round, item),
			)
			if err != nil {
				t.Fatal(err)
			}
		}
		cache.Close()
	}
}

func plaintext(round, item int) string {
	return fmt.Sprintf("Plaintext for round %02d, item %02d", round, item)
}

func ciphertext(round, item int) []byte {
	return []byte(fmt.Sprintf("Ciphertext for round %02d, item %02d", round, item))
}

func encryptNonExistent(t *testing.T, cache *Cache) {
	ciphertext, ok, err := cache.Encrypt("non-existent")
	if err != nil {
		t.Error(err.Error())
	}
	if ok {
		t.Errorf("Encrypting non-existent plaintext returns true")
	}
	if len(ciphertext) != 0 {
		t.Errorf("Encrypting non-existent plaintext returns non-empty ciphertext %x", ciphertext)
	}
}

func decryptNonExistent(t *testing.T, cache *Cache) {
	plaintext, ok, err := cache.Decrypt([]byte("non-existent"))
	if err != nil {
		t.Error(err.Error())
	}
	if ok {
		t.Errorf("Encrypting non-existent ciphertext returns true")
	}
	if len(plaintext) != 0 {
		t.Errorf("Encrypting non-existent ciphertext plaintext returns non-empty plaintext %s", plaintext)
	}
}
