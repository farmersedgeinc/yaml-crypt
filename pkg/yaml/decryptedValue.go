package yaml

import (
	"gopkg.in/yaml.v3"
	"os"
	"github.com/farmersedgeinc/yaml-crypt/pkg/crypto"
)

const decryptedTag = "!secret"

type DecryptedValue struct {
	Value string
	Node *yaml.Node
	Tag bool
}

func (decryptedValue *DecryptedValue) UnmarshalYAML(node *yaml.Node) error {
	var value string
	err := node.Decode(&value)
	if err != nil {
		return err
	}
	decryptedValue.Value = value
	decryptedValue.Node = node
	decryptedValue.Tag = true
	return nil
}

func (decryptedValue *DecryptedValue) toNode() *yaml.Node {
	n := yaml.Node{}
	n.Encode(decryptedValue.Value)
	n.Style = yaml.FlowStyle
	if decryptedValue.Tag {
		n.Tag = decryptedTag
	}
	return &n
}

func (decryptedValue *DecryptedValue) Encrypt(provider crypto.Provider) (*EncryptedValue, error) {
	plainValue := decryptedValue.Value
	data, err := provider.Encrypt(plainValue)
	if err != nil {
		return nil, err
	}
	salt, err := crypto.Salt()
	if err != nil {
		return nil, err
	}
	out := EncryptedValue{
		Salt: salt,
		Hash: crypto.Hash(salt, decryptedValue.Value),
		Data: data,
		Node: decryptedValue.Node,
	}
	return &out, nil
}

func (value *DecryptedValue) ReplaceNode() {
	value.Node.Encode(value.toNode())
}

func getDecryptedValues(node *yaml.Node) (map[string] *DecryptedValue, error) {
	out := map[string] *DecryptedValue{}
	for node := range recursiveNodeIter(node) {
		if node.YamlNode.Tag == decryptedTag {
			var value DecryptedValue
			err := node.YamlNode.Decode(&value)
			if err != nil {
				return out, err
			}
			out[node.Path.String()] = &value
		}
	}
	return out, nil
}

func ReadDecryptedFile(path string) (node *yaml.Node, values map[string] *DecryptedValue, err error) {
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
	values, err = getDecryptedValues(node)
	return
}
