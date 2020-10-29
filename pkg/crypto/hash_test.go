package crypto

import (
	"testing"
	"encoding/hex"
	"reflect"
)

func TestSalt(t *testing.T) {
	for i := 0; i < 1000; i++ {
		salt, err := Salt()
		if err != nil {
			t.Errorf("Salt() failed with error: %s", err.Error())
			i=1000
		}
		if len(salt) != 16 {
			t.Errorf("Salt()'s output was not 16 bytes long. Actual value: %s", hex.EncodeToString(salt))
		}
	}
}

func TestHash(t *testing.T) {
	hash := Hash([]byte("bbbbbbbbbbbbbbbb"), "string")
	shouldBe, _ := hex.DecodeString("731d1c4061ea6ac3b81eb3b302119f67")
	if !reflect.DeepEqual(hash, shouldBe) {
		t.Errorf("Hash()'s output was not 731d1c4061ea6ac3b81eb3b302119f67. Actual value: %s'", hex.EncodeToString(hash))
	}
}
