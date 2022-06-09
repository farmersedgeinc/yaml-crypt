package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/farmersedgeinc/yaml-crypt/pkg/fixtures"
)

func TestEncryptDecryptValue(t *testing.T) {
	repos, err := fixtures.Repos()
	if err != nil {
		t.Fatal(err)
	}
	for _, repo := range repos {
		err := repo.Setup()
		defer repo.Destroy()
		if err != nil {
			t.Fatal(err)
		}
		for _, plaintext := range fixtures.Strings {
			// Round trip via `yaml-crypt encrypt-value | yaml-crypt decrypt-value`
			if !strings.Contains(plaintext, "\n") {
				t.Run(fmt.Sprintf("round trip through yaml-crypt encrypt-value string %q", plaintext), func(t *testing.T) {
					ciphertext, err := encryptValueTest(plaintext+"\n", false)
					if err != nil {
						t.Fatal(fmt.Errorf("Error running yaml-crypt encrypt-value on string %q: %w", plaintext, err))
					}
					newPlaintext, err := decryptValueTest(ciphertext)
					// strip off trailing newline added by command
					newPlaintext = newPlaintext[0 : len(newPlaintext)-1]
					if err != nil {
						t.Fatal(fmt.Errorf("Error running yaml-crypt decrypt-value on string %q (plaintext %q): %w", ciphertext, plaintext, err))
					}
					if plaintext != newPlaintext {
						t.Fatal(fmt.Errorf("Round-trip of string %q decrypted to a different value %q (via ciphertext %q)", plaintext, newPlaintext, ciphertext))
					}
				})
			}
			// Round trip via `yaml-crypt encrypt-value --multi-line | yaml-crypt decrypt-value`
			t.Run(fmt.Sprintf("yaml-crypt encrypt-value --multi-line string %q", plaintext), func(t *testing.T) {
				ciphertext, err := encryptValueTest(plaintext, true)
				if err != nil {
					t.Fatal(fmt.Errorf("Error running yaml-crypt encrypt-value --multi-line on string %q: %w", plaintext, err))
				}
				newPlaintext, err := decryptValueTest(ciphertext)
				// strip off trailing newline added by command
				newPlaintext = newPlaintext[0 : len(newPlaintext)-1]
				if err != nil {
					t.Fatal(fmt.Errorf("Error running yaml-crypt decrypt-value on string %q (plaintext %q): %w", ciphertext, plaintext, err))
				}
				if plaintext != newPlaintext {
					t.Fatal(fmt.Errorf("Round-trip of string %q decrypted to a different value %q (via ciphertext %q)", plaintext, newPlaintext, ciphertext))
				}
			})
		}
	}
}

func encryptValueTest(plaintext string, multiline bool) (string, error) {
	plaintextReader := strings.NewReader(plaintext)
	ciphertext := bytes.Buffer{}
	err := EncryptValue(plaintextReader, &ciphertext, multiline)
	return ciphertext.String(), err
}

func decryptValueTest(ciphertext string) (string, error) {
	ciphertextReader := strings.NewReader(ciphertext)
	plaintext := bytes.Buffer{}
	err := DecryptValue(ciphertextReader, &plaintext, false)
	return plaintext.String(), err
}
