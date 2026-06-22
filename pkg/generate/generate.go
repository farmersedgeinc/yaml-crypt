// Package generate produces cryptographically-random secret values for the
// yaml-crypt !generate tag. Values are built from crypto/rand, guarantee at
// least one character from each required class, enforce a minimum strength
// (length) floor, and are checked against a denylist of known-bad passwords.
//
// Generated plaintext is intended to be born in memory, encrypted, and
// discarded: callers must never persist it to the decrypted source or the
// disk cache.
package generate

import (
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

// StrengthFloor is the minimum length any profile may produce. Profiles below
// it are rejected, so a misconfiguration can never silently weaken a secret.
const StrengthFloor = 16

// DefaultProfile is used when !generate is given with no profile name.
const DefaultProfile = "generic-strong"

// maxDenylistAttempts bounds regeneration when a candidate hits the denylist.
// For random secrets at these lengths a hit is astronomically unlikely; the
// bound just guarantees termination.
const maxDenylistAttempts = 16

const (
	lower  = "abcdefghijklmnopqrstuvwxyz"
	upper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits = "0123456789"
	alnum  = lower + upper + digits
	// strongSymbols deliberately omits quotes, backslash, backtick, and space
	// so values are safe to drop into most shells and config formats.
	strongSymbols = "!#$%*+-=?@^_"
	// connSafeSymbols are URI "unreserved" marks — safe, unescaped, inside
	// database connection strings / URLs.
	connSafeSymbols = "-_.~"
)

// Profile describes how to build a value: its length, the full set of
// characters it may contain, and the classes that must each appear at least
// once.
type Profile struct {
	Length  int
	Charset string
	// Classes each contribute one guaranteed character; every class must be a
	// subset of Charset.
	Classes []string
}

// Profiles is the set of named generation profiles exposed via !generate.
var Profiles = map[string]Profile{
	// connection-string-safe charset, alnum complexity guaranteed
	"cloud-sql": {Length: 32, Charset: alnum + connSafeSymbols, Classes: []string{lower, upper, digits}},
	// default: strong mixed-symbol password
	"generic-strong": {Length: 24, Charset: alnum + strongSymbols, Classes: []string{lower, upper, digits, strongSymbols}},
	// long alphanumeric (no symbols) for systems that reject punctuation
	"alnum-long": {Length: 40, Charset: alnum, Classes: []string{lower, upper, digits}},
	// numeric-only secret at the strength floor
	"pin-numeric": {Length: StrengthFloor, Charset: digits, Classes: []string{digits}},
}

// ProfileNames returns the known profile names, sorted, for error messages.
func ProfileNames() []string {
	names := make([]string, 0, len(Profiles))
	for name := range Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// validate guards against a misconfigured profile: it enforces the strength
// floor and that every required class is a non-empty subset of the charset.
func (p Profile) validate() error {
	if p.Length < StrengthFloor {
		return fmt.Errorf("profile length %d is below the strength floor of %d", p.Length, StrengthFloor)
	}
	if len(p.Charset) == 0 {
		return fmt.Errorf("profile has an empty charset")
	}
	if len(p.Classes) > p.Length {
		return fmt.Errorf("profile requires %d classes but length is only %d", len(p.Classes), p.Length)
	}
	for _, class := range p.Classes {
		if len(class) == 0 {
			return fmt.Errorf("profile has an empty character class")
		}
		for _, r := range class {
			if !strings.ContainsRune(p.Charset, r) {
				return fmt.Errorf("character class %q is not a subset of the charset", class)
			}
		}
	}
	return nil
}

// Value generates a secret for the named profile (empty name → DefaultProfile).
// All randomness comes from crypto/rand.
func Value(profileName string) (string, error) {
	if profileName == "" {
		profileName = DefaultProfile
	}
	p, ok := Profiles[profileName]
	if !ok {
		return "", fmt.Errorf("unknown generate profile %q (known: %s)", profileName, strings.Join(ProfileNames(), ", "))
	}
	if err := p.validate(); err != nil {
		return "", fmt.Errorf("profile %q: %w", profileName, err)
	}
	for attempt := 0; attempt < maxDenylistAttempts; attempt++ {
		candidate, err := p.generateOnce()
		if err != nil {
			return "", err
		}
		if !isDenied(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("profile %q: exceeded denylist rejection attempts", profileName)
}

// generateOnce builds a single candidate: one guaranteed char per class, the
// rest drawn from the full charset, then a crypto-random shuffle so the
// guaranteed characters are not positionally predictable.
func (p Profile) generateOnce() (string, error) {
	buf := make([]byte, p.Length)
	i := 0
	for _, class := range p.Classes {
		c, err := randChar(class)
		if err != nil {
			return "", err
		}
		buf[i] = c
		i++
	}
	for ; i < p.Length; i++ {
		c, err := randChar(p.Charset)
		if err != nil {
			return "", err
		}
		buf[i] = c
	}
	// Fisher-Yates shuffle using crypto/rand.
	for j := len(buf) - 1; j > 0; j-- {
		k, err := randIntn(j + 1)
		if err != nil {
			return "", err
		}
		buf[j], buf[k] = buf[k], buf[j]
	}
	return string(buf), nil
}

// randChar returns a uniformly-random byte from set.
func randChar(set string) (byte, error) {
	idx, err := randIntn(len(set))
	if err != nil {
		return 0, err
	}
	return set[idx], nil
}

// randIntn returns a uniformly-random int in [0, n) using crypto/rand.
func randIntn(n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("randIntn: n must be positive, got %d", n)
	}
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, fmt.Errorf("reading from crypto/rand: %w", err)
	}
	return int(v.Int64()), nil
}

//go:embed passwords/bad-passwords.json
var denylistJSON []byte

// denylist is the lower-cased set of known-bad passwords, loaded once at init.
var denylist map[string]struct{}

func init() {
	var entries []string
	if err := json.Unmarshal(denylistJSON, &entries); err != nil {
		// A malformed embedded asset is a build-time/programmer error.
		panic(fmt.Sprintf("generate: parsing embedded denylist: %v", err))
	}
	denylist = make(map[string]struct{}, len(entries))
	for _, e := range entries {
		denylist[strings.ToLower(e)] = struct{}{}
	}
}

// isDenied reports whether value matches a denylisted password (case-insensitive).
func isDenied(value string) bool {
	_, ok := denylist[strings.ToLower(value)]
	return ok
}
