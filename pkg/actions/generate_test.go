package actions_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache/disk"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache/memory"
	"github.com/farmersedgeinc/yaml-crypt/pkg/crypto"
	"github.com/farmersedgeinc/yaml-crypt/pkg/yaml"
)

const generateDoc = `db:
  password: !generate generic-strong
  pin: !generate pin-numeric
`

// writeRepo lays down a decrypted file containing !generate tags and returns
// the File describing it.
func writeRepo(t *testing.T) (dir string, file actions.File) {
	t.Helper()
	dir = t.TempDir()
	file = actions.File{
		EncryptedPath: filepath.Join(dir, "secrets.encrypted.yaml"),
		DecryptedPath: filepath.Join(dir, "secrets.decrypted.yaml"),
		PlainPath:     filepath.Join(dir, "secrets.plain.yaml"),
	}
	if err := os.WriteFile(file.DecryptedPath, []byte(generateDoc), 0600); err != nil {
		t.Fatal(err)
	}
	return
}

func runEncrypt(t *testing.T, file actions.File, noCache bool) error {
	t.Helper()
	var provider crypto.Provider = crypto.NoopProvider{}
	c, err := memory.Setup()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	return actions.Encrypt([]*actions.File{&file}, c, &provider, 4, 1, time.Second, false, noCache)
}

// encryptedValues returns path->value for all !encrypted nodes. With the noop
// provider the ciphertext equals the plaintext, so this is the generated value.
func encryptedValues(t *testing.T, path string) map[string]string {
	t.Helper()
	node, err := yaml.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	values, err := yaml.GetTaggedChildrenValues(&node, yaml.EncryptedTag)
	if err != nil {
		t.Fatal(err)
	}
	return values
}

func TestGenerateRoundTrip(t *testing.T) {
	_, file := writeRepo(t)
	if err := runEncrypt(t, file, true); err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	values := encryptedValues(t, file.EncryptedPath)
	if len(values) != 2 {
		t.Fatalf("expected 2 encrypted values, got %d: %v", len(values), values)
	}

	var sawStrong, sawPin bool
	for _, v := range values {
		switch len(v) {
		case 24: // generic-strong
			sawStrong = true
			if strings.ContainsAny(v, " \t\n") {
				t.Errorf("generic-strong value contains whitespace: %q", v)
			}
		case 16: // pin-numeric
			sawPin = true
			if strings.TrimFunc(v, func(r rune) bool { return r >= '0' && r <= '9' }) != "" {
				t.Errorf("pin-numeric value is not all digits: %q", v)
			}
		default:
			t.Errorf("unexpected generated length %d for %q", len(v), v)
		}
	}
	if !sawStrong || !sawPin {
		t.Errorf("missing a generated value: strong=%v pin=%v", sawStrong, sawPin)
	}

	// the !generate tag must NOT be written back to the decrypted source.
	src, err := os.ReadFile(file.DecryptedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "!generate") {
		t.Errorf("decrypted source lost its !generate tag; plaintext may have leaked:\n%s", src)
	}
	for _, v := range values {
		if strings.Contains(string(src), v) {
			t.Errorf("generated plaintext leaked into the decrypted source file: %q", v)
		}
	}
}

func TestGenerateIsIdempotent(t *testing.T) {
	_, file := writeRepo(t)
	if err := runEncrypt(t, file, true); err != nil {
		t.Fatalf("first encrypt: %v", err)
	}
	first, err := os.ReadFile(file.EncryptedPath)
	if err != nil {
		t.Fatal(err)
	}

	// re-encrypt: existing !encrypted values must be reused verbatim, not regenerated.
	if err := runEncrypt(t, file, true); err != nil {
		t.Fatalf("second encrypt: %v", err)
	}
	second, err := os.ReadFile(file.EncryptedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("encrypted output changed on re-encrypt (not idempotent):\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestGenerateNeverTouchesDiskCache(t *testing.T) {
	dir, file := writeRepo(t)
	if err := runEncrypt(t, file, true); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	cachePath := filepath.Join(dir, disk.CacheDirName)
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Errorf("disk cache %s exists after a --no-cache generate run (err=%v)", cachePath, err)
	}
}

func TestGenerateRequiresNoCache(t *testing.T) {
	_, file := writeRepo(t)
	err := runEncrypt(t, file, false)
	if err == nil {
		t.Fatal("expected encrypt to refuse generating without --no-cache, got nil")
	}
	if !strings.Contains(err.Error(), "no-cache") {
		t.Errorf("error should mention --no-cache, got: %v", err)
	}
	// nothing should have been written to the encrypted file.
	if _, statErr := os.Stat(file.EncryptedPath); !os.IsNotExist(statErr) {
		t.Errorf("encrypted file was written despite the enforcement error")
	}
}
