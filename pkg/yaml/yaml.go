package yaml

import (
	"encoding/base64"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

const (
	EncryptedTag = "!encrypted"
	DecryptedTag = "!secret"
)

// A Channel-based iterator that yields all descendents of a yaml Node.
func recursiveNodeIter(node *yaml.Node) <-chan *yaml.Node {
	out := make(chan *yaml.Node)
	go func() {
		defer close(out)

		var recurse func(*yaml.Node)
		recurse = func(node *yaml.Node) {
			out <- node
			for _, childNode := range node.Content {
				recurse(childNode)
			}
		}
		recurse(node)
	}()
	return out
}

// A Channel-based iterator that yields all descendents of a yaml Node that match a given tag.
func GetTaggedChildren(node *yaml.Node, tag string) <-chan *yaml.Node {
	out := make(chan *yaml.Node)
	go func() {
		defer close(out)
		for node := range recursiveNodeIter(node) {
			if node.Tag == tag {
				out <- node
			}
		}
	}()
	return out
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
		w, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
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

// Turn a yaml Node tagged !encrypted into a yaml Node tagged !secret, by looking up its values in a give mapping of ciphertexts to plaintexts.
func DecryptNode(node *yaml.Node, decryptionMapping *map[string]string, tag bool) error {
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
	plaintext, ok := (*decryptionMapping)[string(ciphertext)]
	if !ok {
		return errors.New("Ciphertext not found in map. This should never happen.")
	}
	// replace the node contents
	node.Encode(plaintext)
	if tag {
		node.Tag = DecryptedTag
	} else {
		node.Tag = ""
	}
	return nil
}

// Turn a yaml Node tagged !secret into a yaml Node tagged !encrypted, looking up its values in a given mapping of plaintexts to ciphertexts.
func EncryptNode(node *yaml.Node, encryptionMapping *map[string]string) error {
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
	ciphertext, ok := (*encryptionMapping)[plaintext]
	if !ok {
		return errors.New("Plaintext not found in map. This should never happen.")
	}
	// replace the node contents
	node.Encode(base64.StdEncoding.EncodeToString([]byte(ciphertext)))
	node.Tag = EncryptedTag
	return nil
}
