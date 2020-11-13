package yaml

import (
	"encoding/base64"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type Value interface {
	yaml.Unmarshaler
	toNode() *yaml.Node
	ReplaceNode()
}

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

func base64DecodeMapValue(source map[string]interface{}, key string) ([]byte, error) {
	var out []byte
	str, ok := source[key].(string)
	if !ok {
		return out, fmt.Errorf("%s: must be string", key)
	}
	bytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return out, err
	}
	out = bytes
	return out, nil
}

func SaveFile(path string, node *yaml.Node) error {
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
	err = e.Encode(node)
	return err
}
