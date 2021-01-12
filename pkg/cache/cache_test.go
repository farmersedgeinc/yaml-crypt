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
	YoungCacheSize = 100000
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
		// get this round's and the 4 previous rounds' items
		for prevRound := round; prevRound >= round-4 && prevRound >= 0; prevRound-- {
			getItems(t, &cache, prevRound, true)
		}
		err = cache.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	// setup cache, check for items from round 1 to make sure they were removed
	cache, err = Setup(config)
	if err != nil {
		t.Fatal(err)
	}
	getItems(t, &cache, 0, false)
	getItems(t, &cache, 1, false)
	getItems(t, &cache, 19, true)
	err = cache.Close()
	if err != nil {
		t.Fatal(err)
	}
}

// generates the plaintext for a particular round/item
func plaintext(round, item int) string {
	return fmt.Sprintf("Plaintext for round %02d, item %02d", round, item)
}

// generates the ciphertext for a particular round/item
func ciphertext(round, item int) []byte {
	return []byte(fmt.Sprintf("Ciphertext for round %02d, item %02d", round, item))
}

// generates the ciphertext for a particular round/item/version
func versionedCiphertext(round, item, version int) []byte {
	return []byte(fmt.Sprintf("%s, version %d", string(ciphertext(round, item)), version))
}

// put items into the cache for a given round
func putItems(t *testing.T, cache *Cache, round int) {
	for item := 0; item < 100; item++ {
		for version := 0; version < 3; version++ {
			err := cache.Add(
				plaintext(round, item),
				versionedCiphertext(round, item, version),
			)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

// retrieve items from the cache associated with a given round
func getItems(t *testing.T, cache *Cache, round int, shouldSucceed bool) {
	for item := 0; item < 100; item++ {
		for version := 0; item < 3; item++ {
			ct, ok, err := cache.Encrypt(plaintext(round, item), versionedCiphertext(round, item, version))
			if err != nil {
				t.Error(err.Error())
			}
			if shouldSucceed && !ok {
				t.Errorf("Entry not found in cache when encrypting %s", strconv.Quote(plaintext(round, item)))
			} else if !shouldSucceed && ok {
				t.Errorf("Entry should not exist, but found in cache when encrypting %s", strconv.Quote(plaintext(round, item)))
			}
			if ok && shouldSucceed && !bytes.Equal(ct, versionedCiphertext(round, item, version)) {
				t.Errorf("Lookup returned incorrect value when encrypting %s", strconv.Quote(plaintext(round, item)))
			} else if !shouldSucceed && len(ct) > 0 {
				t.Errorf("Entry should not exist, but encrypting returned non-empty []bytes when encrypting %s", strconv.Quote(plaintext(round, item)))
			}

			pt, ok, err := cache.Decrypt(versionedCiphertext(round, item, version))
			if err != nil {
				t.Error(err.Error())
			}
			if shouldSucceed && !ok {
				t.Errorf("Entry not found in cache when decrypting %s", strconv.Quote(string(versionedCiphertext(round, item, version))))
			} else if !shouldSucceed && ok {
				t.Errorf("Entry should not exist, but found in cache when decrypting %s", strconv.Quote(string(versionedCiphertext(round, item, version))))
			}
			if ok && shouldSucceed && pt != plaintext(round, item) {
				t.Errorf("Lookup returned incorrect value when decrypting %s", strconv.Quote(string(versionedCiphertext(round, item, version))))
			} else if !shouldSucceed && len(pt) > 0 {
				t.Errorf("Entry should not exist, but encrypting returned non-empty []bytes when encrypting %s", strconv.Quote(string(versionedCiphertext(round, item, version))))
			}
		}
		if shouldSucceed {
			// try encrypting with an invalid possibleCiphertext. Result should be an arbitrary valid ciphertext.
			ct, ok, err := cache.Encrypt(plaintext(round, item), []byte("invalid ciphertext"))
			if err != nil {
				t.Error(err.Error())
			}
			if !ok {
				t.Errorf("Entry not found in cache when encrypting %s", strconv.Quote(plaintext(round, item)))
			} else if !bytes.HasPrefix(ct, ciphertext(round, item)) {
				t.Errorf("Calling Encrypt() with invalid ciphertext gave invalid result when encrypting %s", strconv.Quote(plaintext(round, item)))
			}
		}
	}
}
