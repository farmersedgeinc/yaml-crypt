package crypto

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func ordinalSuffix(i int) string {
	digit := i % 10
	if digit == 1 {
		return "st"
	} else if digit == 2 {
		return "nd"
	} else if digit == 3 {
		return "rd"
	} else {
		return "th"
	}
}

func TestSalt(t *testing.T) {
	salts := map[string]bool{}
	for i := 0; i < 100000; i++ {
		salt, err := Salt()
		if err != nil {
			t.Errorf("Salt() failed with error: %s", err.Error())
			i = 1000
		}
		salts[string(salt)] = true
		if len(salts) < i+1 {
			t.Errorf("%d%s hash is not unique. Value: %x", i+1, ordinalSuffix(i+1), salt)
		}
		if len(salt) != 16 {
			t.Errorf("Salt()'s output was not 16 bytes long. Actual value: %x", salt)
		}
	}
}

func TestHash(t *testing.T) {
	hash := Hash([]byte("bbbbbbbbbbbbbbbb"), "string")
	shouldBe, _ := hex.DecodeString("731d1c4061ea6ac3b81eb3b302119f67")
	if !reflect.DeepEqual(hash, shouldBe) {
		t.Errorf("Hash()'s output was not 731d1c4061ea6ac3b81eb3b302119f67. Actual value: %x'", hash)
	}
}
