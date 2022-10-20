package yaml

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"gopkg.in/yaml.v3"
)

const (
	EncryptedTag = "!encrypted"
	DecryptedTag = "!secret"
)

// these relations need to be stored to produce "paths" for encrypted values, which is needed for encrypted item reuse
type nodeNode struct {
	YamlNode *yaml.Node
	Path     *Path
}

func recursiveNodeIter(node *yaml.Node) <-chan *nodeNode {
	out := make(chan *nodeNode)
	go func() {
		defer close(out)

		var recurse func(*yaml.Node, *nodeNode, int)
		recurse = func(node *yaml.Node, parent *nodeNode, index int) {
			var path *Path
			if parent != nil {
				if parent.YamlNode.Kind == yaml.MappingNode {
					if index > 0 && index%2 == 1 {
						path = parent.Path.AddString(parent.YamlNode.Content[index-1].Value)
					}
				} else {
					path = parent.Path.AddInt(index)
				}
			} else {
				path = &Path{isInt: true, i: index}
			}
			current := &nodeNode{YamlNode: node, Path: path}

			out <- current
			parent = current
			for index, childNode := range node.Content {
				recurse(childNode, current, index)
			}
		}
		recurse(node, nil, 0)
	}()
	return out
}

func DeepCopyNode(node *yaml.Node) *yaml.Node {
	result := *node
	result.Content = nil
	for _, item := range node.Content {
		result.Content = append(result.Content, DeepCopyNode(item))
	}
	if node.Alias != nil {
		result.Alias = DeepCopyNode(node.Alias)
	}
	return &result
}

// A Channel-based iterator that yields all descendents of a yaml Node that match a given tag.
func GetTaggedChildren(node *yaml.Node, tag string) <-chan *nodeNode {
	out := make(chan *nodeNode)
	go func() {
		defer close(out)
		for node := range recursiveNodeIter(node) {
			if node.YamlNode.Tag == tag {
				out <- node
			}
		}
	}()
	return out
}

// Get a map of paths to decoded string values from all descendents of a yaml Node that match a given tag.
func GetTaggedChildrenValues(node *yaml.Node, tag string) (out map[string]string, err error) {
	out = map[string]string{}
	for n := range GetTaggedChildren(node, tag) {
		var value string
		value, err = GetValue(n.YamlNode)
		if err != nil {
			return
		}
		out[n.Path.String()] = value
	}
	return
}

// Read a yaml file, and return its root yaml Node.
func ReadFile(path string) (node yaml.Node, err error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return
	}
	err = yaml.NewDecoder(f).Decode(&node)
	return
}

// Save a yaml Node to a file.
func SaveFile(path string, node yaml.Node) error {
	var w io.Writer
	var err error
	if path == "" {
		w = os.Stdout
	} else {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		err = f.Truncate(0)
		if err != nil {
			return err
		}
		w = f
	}
	e := yaml.NewEncoder(w)
	e.SetIndent(2)
	err = e.Encode(&node)
	return err
}

// Get the decoded value of an !encrypted or !secret Node, as a String. !encrypted Nodes are base64-decoded.
func GetValue(node *yaml.Node) (value string, err error) {
	if node.Tag == EncryptedTag {
		var encodedCiphertext string
		err = node.Decode(&encodedCiphertext)
		if err != nil {
			return
		}
		var bytes []byte
		bytes, err = base64.StdEncoding.DecodeString(encodedCiphertext)
		value = string(bytes)
	} else if node.Tag == DecryptedTag {
		err = node.Decode(&value)
	} else {
		err = fmt.Errorf("Node must be tagged %s or %s", EncryptedTag, DecryptedTag)
	}
	return
}

// Strip the tag from any children of a yaml Node that are tagged with the given tag.
func StripTags(node *yaml.Node, tag string) {
	for child := range GetTaggedChildren(node, tag) {
		child.YamlNode.Tag = ""
	}
}

// Turn a yaml Node tagged !encrypted into a yaml Node tagged !secret, by looking up its values in a give mapping of ciphertexts to plaintexts.
func DecryptNode(node *yaml.Node, cache *cache.Cache) error {
	// validate, read in data
	if node.Tag != EncryptedTag {
		return fmt.Errorf("Cannot decrypt a node not tagged %s", EncryptedTag)
	}
	var encodedCiphertext string
	err := node.Decode(&encodedCiphertext)
	if err != nil {
		return err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return err
	}
	// decrypt
	plaintext, ok, err := cache.Decrypt(ciphertext)
	if err != nil {
		return err
	} else if !ok {
		return errors.New("Ciphertext not found in cache. This should never happen.")
	}
	// replace the node contents
	node.Encode(plaintext)
	node.Tag = DecryptedTag
	return nil
}

// Turn a yaml Node tagged !secret into a yaml Node tagged !encrypted, looking up its values in a given mapping of plaintexts to ciphertexts.
func EncryptNode(node *yaml.Node, possibleCiphertext []byte, cache *cache.Cache) error {
	// validate, read in data
	if node.Tag != DecryptedTag {
		return fmt.Errorf("Cannot encrypt a node not tagged %s", DecryptedTag)
	}
	var plaintext string
	err := node.Decode(&plaintext)
	if err != nil {
		return err
	}
	// encrypt
	ciphertext, ok, err := cache.Encrypt(plaintext, possibleCiphertext)
	if !ok {
		return errors.New("Plaintext not found in cache. This should never happen.")
	}
	// replace the node contents
	node.Encode(base64.StdEncoding.EncodeToString([]byte(ciphertext)))
	node.Tag = EncryptedTag
	return nil
}
