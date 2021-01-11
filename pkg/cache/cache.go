package cache

import (
	"crypto/sha256"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/prologic/bitcask"
	"os"
	"path/filepath"
	"sync"
)

const (
	// Length to hash plaintext and ciphertext keys.
	hashLength = 16
	// Prefix for keys containing a hashed plaintext, used to look up ciphertext.
	plaintextKeyPrefix = 'p'
	// Prefix for keys containing a hashed ciphertext, used to look up plaintext.
	ciphertextKeyPrefix = 'c'
	// Name of the directory to store the caches in
	CacheDirName = ".yamlcrypt.cache"
)

// Max young cache size: 100MiB by default (can be shrunk for tests)
var YoungCacheSize int64 = 1024 * 1024 * 100

// A quick and dirty "LRU-ish" cache.
// Maintains a read/write "young" cache, and a read-only "old" cache.
// New values are added to the "young" cache.
// When looking up a value, if it's present in the "young" cache, retrieve it from there. If it's present in the "old" cache, retrieve it from there, copying it into the "young" cache.
// When the "young" cache gets too big, the current "old" cache is removed and the current "young" cache takes its place. This only happens on close since the lifecycle of this object is expected to be pretty short in this application, but the benefit of this is: during a session, any values added to the cache are guaranteed to remain present until at least the end of the session (technically, until the end of the next session, due to the "old" cache).
// Getting and inserting values are protected with a mutex, making this safe for parallel access, if a bit of a drag.
type Cache struct {
	young     *bitcask.Bitcask
	youngPath string
	old       *bitcask.Bitcask
	oldPath   string
	mutex     sync.Mutex
}

// Initialize the cache.
func Setup(config config.Config) (Cache, error) {
	cache := Cache{
		youngPath: filepath.Join(config.Root, CacheDirName, "young"),
		oldPath:   filepath.Join(config.Root, CacheDirName, "old"),
	}
	var err error
	cache.young, err = bitcask.Open(
		cache.youngPath,
		bitcask.WithAutoRecovery(true),
	)
	if err != nil {
		return cache, err
	}
	cache.old, err = bitcask.Open(
		cache.oldPath,
		bitcask.WithAutoRecovery(true),
	)
	return cache, err
}

// Close the cache, doing some cleanup as well. Must be called before exiting
func (c *Cache) Close() error {
	// we only need to merge young, because old is read-only
	mergeErr := c.young.Merge()
	stats, statsErr := c.young.Stats()
	// we want to close if at all possible, so we'll handle merge/stats errors later
	err := c.young.Close()
	if err != nil {
		return err
	}
	err = c.old.Close()
	if err != nil {
		return err
	}
	if mergeErr != nil {
		return mergeErr
	}
	if statsErr != nil {
		return statsErr
	}
	// if the young cache size is too big, get rid of the old cache and make the young cache take its place.
	if stats.Size > YoungCacheSize {
		err := os.RemoveAll(c.oldPath)
		if err != nil {
			return err
		}
		return os.Rename(c.youngPath, c.oldPath)
	}
	return nil
}

// Look up the ciphertext for a given plaintext. Protected with a mutex.
func (c *Cache) Encrypt(plaintext string, potentialCiphertext []byte) (ciphertext []byte, ok bool, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var set CiphertextSet
	set, ok, err = c.get(plaintextToKey(plaintext))
	if err != nil {
		return
	}
	if len(set) == 0 {
	} else if len(potentialCiphertext) > 0 && set.Lookup(potentialCiphertext) {
		ciphertext = potentialCiphertext
		ok = true
	} else {
		ciphertext, err = set.GetOne()
		ok = len(ciphertext) > 0
	}
	return
}

// Look up the plaintext for a given ciphertext. Protected with a mutex.
func (c *Cache) Decrypt(ciphertext []byte) (string, bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	plaintext, ok, err := c.get(ciphertextToKey(ciphertext))
	return string(plaintext), ok, err
}

// Add a (plaintext, ciphertext) pair to the young cache. Protected with a mutex.
func (c *Cache) Add(plaintext string, ciphertext []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.add(plaintext, ciphertext)
}

// Add a (plaintext, ciphertext) pair to the young cache.
func (c *Cache) add(plaintext string, ciphertext []byte) error {
	var set CiphertextSet
	plaintextKey := plaintextToKey(plaintext)
	set, _, err := c.get(plaintextKey)
	if err != nil {
		return err
	}
	err = c.young.Put(plaintextKey, set.Add(ciphertext))
	if err != nil {
		return err
	}
	return c.young.Put(ciphertextToKey(ciphertext), []byte(plaintext))
}

func (c *Cache) get(key []byte) (value []byte, ok bool, err error) {
	if c.young.Has(key) {
		value, err = c.young.Get(key)
		ok = true
	} else if c.old.Has(key) {
		value, err = c.old.Get(key)
		if err != nil {
			return
		}
		ok = true
		err = c.young.Put(key, value)
	}
	return
}

// Convert a ciphertext to the key used to lookup its plaintext.
func ciphertextToKey(data []byte) []byte {
	key := make([]byte, 1, hashLength+1)
	key[0] = ciphertextKeyPrefix
	key = append(key, hash(data)...)
	return key
}

// Convert a plaintext to the key used to lookup its ciphertext.
func plaintextToKey(data string) []byte {
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
