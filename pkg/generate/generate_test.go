package generate

import (
	"strings"
	"testing"
)

// classOf reports which of the four base classes a byte belongs to.
func classes(s string) (hasLower, hasUpper, hasDigit, hasSymbol bool) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			hasSymbol = true
		}
	}
	return
}

// TestProfilesMeetStrengthFloor guards against a misconfigured profile silently
// weakening a secret: every profile must validate and meet the length floor.
func TestProfilesMeetStrengthFloor(t *testing.T) {
	for name, p := range Profiles {
		if err := p.validate(); err != nil {
			t.Errorf("profile %q failed validation: %v", name, err)
		}
		if p.Length < StrengthFloor {
			t.Errorf("profile %q length %d is below strength floor %d", name, p.Length, StrengthFloor)
		}
	}
}

// TestGeneratePropertiesPerProfile checks length, charset membership, and the
// at-least-one-per-class guarantee, over many samples per profile.
func TestGeneratePropertiesPerProfile(t *testing.T) {
	for name, p := range Profiles {
		for n := 0; n < 200; n++ {
			v, err := Value(name)
			if err != nil {
				t.Fatalf("profile %q: unexpected error: %v", name, err)
			}
			if len(v) != p.Length {
				t.Fatalf("profile %q: got length %d, want %d", name, len(v), p.Length)
			}
			for i := 0; i < len(v); i++ {
				if !strings.ContainsRune(p.Charset, rune(v[i])) {
					t.Fatalf("profile %q: char %q not in charset", name, v[i])
				}
			}
			// every required class must appear at least once
			for _, class := range p.Classes {
				found := false
				for i := 0; i < len(v); i++ {
					if strings.ContainsRune(class, rune(v[i])) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("profile %q: no character from required class %q in %q", name, class, v)
				}
			}
		}
	}
}

func TestPinNumericIsAllDigits(t *testing.T) {
	for n := 0; n < 100; n++ {
		v, err := Value("pin-numeric")
		if err != nil {
			t.Fatal(err)
		}
		_, _, hasDigit, hasSymbol := classes(v)
		if !hasDigit || hasSymbol || strings.ContainsAny(v, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			t.Fatalf("pin-numeric produced non-numeric value %q", v)
		}
	}
}

func TestGenericStrongHasAllClasses(t *testing.T) {
	v, err := Value("generic-strong")
	if err != nil {
		t.Fatal(err)
	}
	hasLower, hasUpper, hasDigit, hasSymbol := classes(v)
	if !(hasLower && hasUpper && hasDigit && hasSymbol) {
		t.Fatalf("generic-strong missing a class: lower=%v upper=%v digit=%v symbol=%v (%q)",
			hasLower, hasUpper, hasDigit, hasSymbol, v)
	}
}

func TestEmptyProfileUsesDefault(t *testing.T) {
	v, err := Value("")
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != Profiles[DefaultProfile].Length {
		t.Fatalf("empty profile: got length %d, want default %d", len(v), Profiles[DefaultProfile].Length)
	}
}

func TestUnknownProfileErrors(t *testing.T) {
	if _, err := Value("does-not-exist"); err == nil {
		t.Fatal("expected error for unknown profile, got nil")
	}
}

func TestValuesAreDistinct(t *testing.T) {
	seen := map[string]struct{}{}
	for n := 0; n < 100; n++ {
		v, err := Value("generic-strong")
		if err != nil {
			t.Fatal(err)
		}
		if _, dup := seen[v]; dup {
			t.Fatalf("crypto/rand produced a duplicate value %q within 100 draws", v)
		}
		seen[v] = struct{}{}
	}
}

func TestDenylistRejectsKnownBad(t *testing.T) {
	if !isDenied("password") {
		t.Error("expected 'password' to be denied")
	}
	if !isDenied("PASSWORD") {
		t.Error("denylist should be case-insensitive")
	}
	if isDenied("aЗ9-not-in-list-xyz") {
		t.Error("did not expect a random value to be denied")
	}
}
