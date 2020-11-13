package crypto

import (
	"context"
	"github.com/farmersedgeinc/yaml-crypt/pkg/fixtures"
	"golang.org/x/oauth2/google"
	"reflect"
	"strconv"
	"testing"
)

type ProviderMeta struct {
	Provider Provider
	Skip     func() bool
}

var providers = []ProviderMeta{
	ProviderMeta{NoopProvider{}, func() bool { return false }},
	ProviderMeta{
		GoogleProvider{
			Project:  "yaml-crypt-test-9420f5b24e736f",
			Location: "global",
			Keyring:  "yamlcrypt-test-49809b3c30a6e22",
			Key:      "yaml-crypt",
		},
		func() bool {
			_, err := google.FindDefaultCredentials(context.Background())
			return err != nil
		},
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
				ciphertext, err := provider.Encrypt(original)
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
				plaintext, err := provider.Decrypt(ciphertext)
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
