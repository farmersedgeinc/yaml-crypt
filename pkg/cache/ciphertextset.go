package cache

import (
	"bytes"
	"encoding/base64"
)

type CiphertextSet []byte

func (set CiphertextSet) Lookup(ciphertext []byte) bool {
	encodedQuery := []byte(base64.StdEncoding.EncodeToString(ciphertext))
	for _, encodedItem := range bytes.Split(set, []byte(" ")) {
		if bytes.Equal(encodedItem, encodedQuery) {
			return true
		}
	}
	return false
}

func (set CiphertextSet) Add(ciphertext []byte) CiphertextSet {
	encoded := CiphertextSet(base64.StdEncoding.EncodeToString(ciphertext))
	if len(set) == 0 {
		return encoded
	}
	ok := set.Lookup(ciphertext)
	if ok {
		return set
	}
	return CiphertextSet(string(ciphertext) + " " + string(encoded))
}

func (set CiphertextSet) GetOne() ([]byte, error) {
	items := bytes.Split(set, []byte(" "))
	if len(items) == 0 {
		return []byte{}, nil
	}
	return base64.StdEncoding.DecodeString(string(items[0]))
}
