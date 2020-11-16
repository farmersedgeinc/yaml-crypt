package cache

import (
	"bytes"
	"fmt"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/fixtures"
	"strconv"
	"testing"
)

func TestCache(t *testing.T) {
	// make the cache a lot smaller to make it quicker to test LRU behavior
	YoungCacheSize = 1024 * 100
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
	// setup cache, check for non-existent items
	cache, err := Setup(config)
	if err != nil {
		t.Fatal(err)
	}
	getItems(t, &cache, 0, false)
	err = cache.Close()
	if err != nil {
		t.Fatal(err)
	}
	// 20 rounds, enough to trigger multiple cache cleanups:
	for round := 0; round < 20; round++ {
		cache, err := Setup(config)
		if err != nil {
			t.Fatal(err)
		}
		// put this round's items
		putItems(t, &cache, round)
		// get this round's and the 2 previous rounds' items
		for prevRound := round; prevRound >= round-2 && prevRound >= 0; prevRound-- {
			getItems(t, &cache, round, true)
		}
		err = cache.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	// setup cache, check for items from round 1, which should be gone due to the cleanup
	cache, err = Setup(config)
	if err != nil {
		t.Fatal(err)
	}
	getItems(t, &cache, 0, false)
	err = cache.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func plaintext(round, item int) string {
	return fmt.Sprintf("Plaintext for round %02d, item %02d", round, item)
}

func ciphertext(round, item int) []byte {
	return []byte(fmt.Sprintf("Ciphertext for round %02d, item %02d", round, item))
}

func putItems(t *testing.T, cache *Cache, round int) {
	for item := 0; item < 100; item++ {
		err := cache.Add(
			plaintext(round, item),
			ciphertext(round, item),
		)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func getItems(t *testing.T, cache *Cache, round int, shouldSucceed bool) {
	for item := 0; item < 100; item++ {
		ct, ok, err := cache.Encrypt(plaintext(round, item))
		if err != nil {
			t.Error(err.Error())
		}
		if shouldSucceed && !ok {
			t.Errorf("Entry not found in cache when encrypting %s", strconv.Quote(plaintext(round, item)))
		} else if !shouldSucceed && ok {
			t.Errorf("Entry should not exist, but found in cache when encrypting %s", strconv.Quote(plaintext(round, item)))
		}
		if ok && shouldSucceed && !bytes.Equal(ct, ciphertext(round, item)) {
			t.Errorf("Lookup returned incorrect value when encrypting %s", strconv.Quote(plaintext(round, item)))
		} else if !shouldSucceed && len(ct) > 0 {
			t.Errorf("Entry should not exist, but encrypting returned non-empty []bytes when encrypting %s", strconv.Quote(plaintext(round, item)))
		}

		pt, ok, err := cache.Decrypt(ciphertext(round, item))
		if err != nil {
			t.Error(err.Error())
		}
		if shouldSucceed && !ok {
			t.Errorf("Entry not found in cache when decrypting %s", strconv.Quote(string(ciphertext(round, item))))
		} else if !shouldSucceed && ok {
			t.Errorf("Entry should not exist, but found in cache when decrypting %s", strconv.Quote(string(ciphertext(round, item))))
		}
		if ok && shouldSucceed && pt != plaintext(round, item) {
			t.Errorf("Lookup returned incorrect value when decrypting %s", strconv.Quote(string(ciphertext(round, item))))
		} else if !shouldSucceed && len(pt) > 0 {
			t.Errorf("Entry should not exist, but encrypting returned non-empty []bytes when encrypting %s", strconv.Quote(string(ciphertext(round, item))))
		}
	}
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
