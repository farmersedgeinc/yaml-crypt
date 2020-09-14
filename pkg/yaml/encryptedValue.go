package yaml

import (
	"gopkg.in/yaml.v3"
	"os"
	"encoding/base64"
	"bytes"
	"github.com/farmersedgeinc/yaml-crypt/pkg/crypto"
)

const encryptedTag = "!encrypted"

type EncryptedValue struct {
	Salt []byte
	Hash []byte
	Data []byte
	Node *yaml.Node
}

func (encryptedValue *EncryptedValue) UnmarshalYAML(node *yaml.Node) error {
	var value map[string]interface{}
	err := node.Decode(&value)
	if err != nil {
		return err
	}
	encryptedValue.Salt, err = base64DecodeMapValue(value, "salt")
	if err != nil {
		return err
	}
	encryptedValue.Hash, err = base64DecodeMapValue(value, "hash")
	if err != nil {
		return err
	}
	encryptedValue.Data, err = base64DecodeMapValue(value, "data")
	if err != nil {
		return err
	}
	encryptedValue.Node = node
	return nil
}

func (encryptedValue *EncryptedValue) toNode() *yaml.Node {
	type tmp struct {
		Salt string
		Hash string
		Data string
	}
	n := yaml.Node{}
	n.Encode(tmp{
		base64.StdEncoding.EncodeToString(encryptedValue.Salt),
		base64.StdEncoding.EncodeToString(encryptedValue.Hash),
		base64.StdEncoding.EncodeToString(encryptedValue.Data),
	})
	n.Style = yaml.FlowStyle
	n.Tag = encryptedTag
	return &n
}

func getEncryptedValues(node *yaml.Node) (map[string] *EncryptedValue, error) {
	out := map[string] *EncryptedValue{}
	for node := range recursiveNodeIter(node) {
		if node.YamlNode.Tag == encryptedTag {
			var value EncryptedValue
			err := node.YamlNode.Decode(&value)
			if err != nil {
				return out, err
			}
			out[node.Path.String()] = &value
		}
	}
	return out, nil
}

func ReadEncryptedFile(path string) (node *yaml.Node, values map[string] *EncryptedValue, err error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return
	}
	var n yaml.Node
	err = yaml.NewDecoder(f).Decode(&n)
	if err != nil {
		return
	}
	node = &n
	values, err = getEncryptedValues(node)
	return
}

func (encryptedValue *EncryptedValue) Compare(value *DecryptedValue) bool {
	salt := encryptedValue.Salt
	return bytes.Equal(encryptedValue.Hash, crypto.Hash(salt, value.Value))
}

func (encryptedValue *EncryptedValue) Decrypt(provider crypto.Provider, tag bool) (*DecryptedValue, error) {
	value, err := provider.Decrypt(encryptedValue.Data)
	if err != nil {
		return nil, err
	}
	return &DecryptedValue {
		Value: value,
		Node: encryptedValue.Node,
		Tag: tag,
	}, nil
}

func (value *EncryptedValue) ReplaceNode() {
	value.Node.Encode(value.toNode())
}
