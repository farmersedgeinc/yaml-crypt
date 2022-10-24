package crypto

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/farmersedgeinc/yaml-crypt/pkg/fixtures"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type ProviderMeta struct {
	Provider        Provider
	InvalidProvider Provider
	Skip            func() bool
	HasInvalid      bool
}

const retries = 5
const timeout = 10 * time.Second

var providers = []ProviderMeta{
	ProviderMeta{NoopProvider{}, NoopProvider{}, func() bool { return false }, false},
	ProviderMeta{
		GoogleProvider{
			Project:  "yaml-crypt-test-9420f5b24e736f",
			Location: "global",
			Keyring:  "yamlcrypt-test-49809b3c30a6e22",
			Key:      "yaml-crypt",
		},
		GoogleProvider{
			Project:  "yaml-crypt-test-9420f5b24e736f",
			Location: "global",
			Keyring:  "yamlcrypt-test-49809b3c30a6e22",
			Key:      "yaml-crypt",
			Options:  []option.ClientOption{option.WithCredentialsFile("/dev/null")},
		},
		func() bool {
			_, err := google.FindDefaultCredentials(context.Background())
			return err != nil
		},
		true,
	},
}

func TestRoundTrip(t *testing.T) {
	for _, meta := range providers {
		provider := meta.Provider
		name := reflect.TypeOf(provider).Name()
		t.Run(name, func(t *testing.T) {
			if meta.Skip() {
				t.Skip()
			}
			for _, original := range fixtures.Strings {
				ciphertext, err := provider.Encrypt(original, retries, timeout)
				if err != nil {
					t.Errorf(
						"Provider %s failed to encrypt with error: %s\nPlaintext: %s",
						name,
						strconv.Quote(err.Error()),
						strconv.Quote(original),
					)
				} else if len(ciphertext) == 0 {
					t.Errorf(
						"Provider %s produced zero-length ciphertext\nPlaintext: %s",
						name,
						strconv.Quote(original),
					)
				}
				plaintext, err := provider.Decrypt(ciphertext, retries, timeout)
				if err != nil {
					t.Errorf(
						"Provider %s failed to decrypt with error: %s\nCiphertext: %x\nExpected Plaintext: %s",
						name,
						strconv.Quote(err.Error()),
						ciphertext,
						strconv.Quote(original),
					)
				} else if original != plaintext {
					t.Errorf(
						"Round-trip failed for provider %s\nOriginal Plaintext: %s\nCiphertext: %x\nFinal Plaintext: %s",
						name,
						strconv.Quote(original),
						ciphertext,
						strconv.Quote(plaintext),
					)

				}
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	for _, meta := range providers {
		provider := meta.InvalidProvider
		name := reflect.TypeOf(provider).Name()
		t.Run(name, func(t *testing.T) {
			if meta.Skip() || !meta.HasInvalid {
				t.Skip()
			}
			_, err := provider.Encrypt("test", retries, timeout)
			if err == nil {
				t.Errorf("Provider %s did not fail to encrypt when given invalid configuration", name)
			}
			_, err = provider.Decrypt([]byte("test"), retries, timeout)
			if err == nil {
				t.Errorf("Provider %s did not fail to decrypt when given invalid configuration", name)
			}
		})
	}
}
