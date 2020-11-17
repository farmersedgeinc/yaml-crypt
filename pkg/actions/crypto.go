package actions

import (
	"fmt"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/crypto"
	"github.com/farmersedgeinc/yaml-crypt/pkg/yaml"
	yamlv3 "gopkg.in/yaml.v3"
	"strconv"
)

func Decrypt(f *File, plain bool, stdout bool, cache *cache.Cache, provider *crypto.Provider, threads int) error {
	// read root node from file
	node, err := yaml.ReadFile(f.EncryptedPath)
	if err != nil {
		return err
	}
	// make decryption mapping for encrypted child nodes
	mapping, err := getDecryptionMapping(node, cache, provider, threads)
	if err != nil {
		return err
	}
	// apply decryption mapping to encrypted child nodes
	for node := range yaml.GetTaggedChildren(&node, yaml.EncryptedTag) {
		yaml.DecryptNode(node, &mapping, !plain)
	}
	// write modified root node out to file
	var outPath string
	if stdout {
		outPath = ""
	} else if plain {
		outPath = f.PlainPath
	} else {
		outPath = f.DecryptedPath
	}
	return yaml.SaveFile(outPath, node)
}

// get the decryption mapping for a root node
func getDecryptionMapping(node yamlv3.Node, cache *cache.Cache, provider *crypto.Provider, threads int) (out map[string]string, err error) {
	out = map[string]string{}
	for node := range yaml.GetTaggedChildren(&node, yaml.EncryptedTag) {
		var value string
		value, err = yaml.GetValue(node)
		if err != nil {
			return
		}
		out[value] = ""
	}
	err = fillDecryptionMapping(&out, cache, provider, threads)
	return
}

func printMap(m map[string]string) {
	for k, v := range m {
		fmt.Printf("%s: %s\n", strconv.Quote(k), strconv.Quote(v))
	}
}

func Encrypt(f *File, cache *cache.Cache, provider *crypto.Provider, threads int) error {
	// if there's an encrypted file, decrypt it just to populate the cache
	if exists(f.EncryptedPath) {
		node, err := yaml.ReadFile(f.EncryptedPath)
		if err != nil {
			return err
		}
		_, err = getDecryptionMapping(node, cache, provider, threads)
		if err != nil {
			return err
		}
	}
	// read in the decrypted file
	node, err := yaml.ReadFile(f.DecryptedPath)
	if err != nil {
		return err
	}
	// get encryption mapping
	mapping, err := getEncryptionMapping(node, cache, provider, threads)
	if err != nil {
		return err
	}
	// apply encryption mapping to decrypted child nodes
	for node := range yaml.GetTaggedChildren(&node, yaml.DecryptedTag) {
		yaml.EncryptNode(node, &mapping)
	}
	// write output
	err = yaml.SaveFile(f.EncryptedPath, node)
	return err
}

// get the encryption mapping for a root node
func getEncryptionMapping(node yamlv3.Node, cache *cache.Cache, provider *crypto.Provider, threads int) (out map[string]string, err error) {
	out = map[string]string{}
	for node := range yaml.GetTaggedChildren(&node, yaml.DecryptedTag) {
		var value string
		value, err = yaml.GetValue(node)
		if err != nil {
			return
		}
		out[value] = ""
	}
	err = fillEncryptionMapping(&out, cache, provider, threads)
	return
}

func fillEncryptionMapping(mapping *map[string]string, cache *cache.Cache, provider *crypto.Provider, threads int) error {
	var misses []string
	// get everything we can from the cache
	for plaintext := range *mapping {
		ciphertext, ok, err := cache.Encrypt(plaintext)
		if err != nil {
			return err
		}
		if ok {
			(*mapping)[plaintext] = string(ciphertext)
		} else {
			misses = append(misses, plaintext)
		}
	}
	// encrypt anything that missed the cache
	if len(misses) > 0 {
		m, err := parallelMap(misses, func(plaintext string) (string, error) {
			ciphertext, err := (*provider).Encrypt(plaintext)
			if err != nil {
				return "", err
			}
			err = cache.Add(plaintext, ciphertext)
			return string(ciphertext), err
		}, threads)
		if err != nil {
			return err
		}
		for plaintext, ciphertext := range m {
			(*mapping)[plaintext] = ciphertext
		}
	}
	return nil
}

func fillDecryptionMapping(mapping *map[string]string, cache *cache.Cache, provider *crypto.Provider, threads int) error {
	var misses []string
	// get everything we can from the cache
	for ciphertext := range *mapping {
		plaintext, ok, err := cache.Decrypt([]byte(ciphertext))
		if err != nil {
			return err
		}
		if ok {
			(*mapping)[ciphertext] = plaintext
		} else {
			misses = append(misses, ciphertext)
		}
	}
	// decrypt anything that missed the cache
	if len(misses) > 0 {
		m, err := parallelMap(misses, func(ciphertext string) (string, error) {
			plaintext, err := (*provider).Decrypt([]byte(ciphertext))
			if err != nil {
				return "", err
			}
			err = cache.Add(plaintext, []byte(ciphertext))
			return plaintext, err
		}, threads)
		if err != nil {
			return err
		}
		for ciphertext, plaintext := range m {
			(*mapping)[ciphertext] = plaintext
		}
	}
	return nil
}

func parallelMap(inputs []string, function func(string) (string, error), threads int) (outputs map[string]string, err error) {
	inputChannel := make(chan string)
	outputChannel := make(chan mapResult)
	outputs = map[string]string{}
	// spin up workers
	for i := 0; i < threads; i++ {
		go func() {
			for input := range inputChannel {
				output, err := function(input)
				outputChannel <- mapResult{input, output, err}
			}
		}()
	}
	// feed workers
	go func() {
		for _, input := range inputs {
			inputChannel <- input
		}
	}()
	// consume results
	for i := 0; i < len(inputs); i++ {
		result := <-outputChannel
		err = result.err
		if err != nil {
			return
		}
		outputs[result.input] = result.output
	}
	close(inputChannel)
	close(outputChannel)
	return
}

type mapResult struct {
	input  string
	output string
	err    error
}
