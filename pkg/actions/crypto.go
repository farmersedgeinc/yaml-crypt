package actions

import (
	"fmt"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/crypto"
	"github.com/farmersedgeinc/yaml-crypt/pkg/yaml"
	yamlv3 "gopkg.in/yaml.v3"
	"strconv"
)

func Decrypt(files []*File, plain bool, stdout bool, cache *cache.Cache, provider *crypto.Provider, threads int) (err error) {
	nodes := make([]yamlv3.Node, len(files))
	decryptionMapping := map[string]string{}
	// read in files, set decryption mapping keys
	for i, file := range files {
		nodes[i], err = yaml.ReadFile(file.EncryptedPath)
		if err != nil {
			return
		}
		err = addTaggedValuesToMapping(&decryptionMapping, &nodes[i], yaml.EncryptedTag)
		if err != nil {
			return
		}
	}
	// fill in the values of the decryption mapping
	err = fillDecryptionMapping(&decryptionMapping, cache, provider, threads)
	if err != nil {
		return
	}
	for i, file := range files {
		// apply decryption mapping to encrypted child nodes
		for node := range yaml.GetTaggedChildren(&nodes[i], yaml.EncryptedTag) {
			err = yaml.DecryptNode(node, &decryptionMapping, !plain)
			if err != nil {
				return err
			}
		}
		// write modified root node out to file
		var outPath string
		if stdout {
			outPath = ""
		} else if plain {
			outPath = file.PlainPath
		} else {
			outPath = file.DecryptedPath
		}
		err = yaml.SaveFile(outPath, nodes[i])
		if err != nil {
			return
		}
	}
	return
}

func addTaggedValuesToMapping(mapping *map[string]string, node *yamlv3.Node, tag string) (err error) {
	values, err := yaml.GetTaggedChildrenValues(node, tag)
	if err != nil {
		return
	}
	for _, value := range values {
		(*mapping)[value] = ""
	}
	return
}

func printMap(m map[string]string) {
	for k, v := range m {
		fmt.Printf("%s: %s\n", strconv.Quote(k), strconv.Quote(v))
	}
}

func Encrypt(files []*File, cache *cache.Cache, provider *crypto.Provider, threads int) (err error) {
	nodes := make([]yamlv3.Node, len(files))
	decryptionMapping := map[string]string{}
	encryptionMapping := map[string]string{}
	for i, file := range files {
		// read in decrypted file
		nodes[i], err = yaml.ReadFile(file.DecryptedPath)
		if err != nil {
			return
		}
		err = addTaggedValuesToMapping(&encryptionMapping, &nodes[i], yaml.DecryptedTag)
		if err != nil {
			return
		}
		// if an encrypted version exists, load its encrypted values and add them to the decryption mapping, for cache stuff later
		if exists(file.EncryptedPath) {
			var node yamlv3.Node
			node, err = yaml.ReadFile(file.EncryptedPath)
			if err != nil {
				return
			}
			err = addTaggedValuesToMapping(&decryptionMapping, &node, yaml.EncryptedTag)
			if err != nil {
				return
			}
		}
	}
	// decrypt any encrypted values first, to pre-fill the cache with their existing versions
	err = fillDecryptionMapping(&decryptionMapping, cache, provider, threads)
	if err != nil {
		return
	}
	// now we can fill in the values of the encryption mapping, with any exising values coming from the cache
	err = fillEncryptionMapping(&encryptionMapping, cache, provider, threads)
	if err != nil {
		return
	}

	for i, file := range files {
		// apply encryption mapping to decrypted child nodes
		for node := range yaml.GetTaggedChildren(&nodes[i], yaml.DecryptedTag) {
			err = yaml.EncryptNode(node, &encryptionMapping)
			if err != nil {
				return
			}
		}
		// write output
		err = yaml.SaveFile(file.EncryptedPath, nodes[i])
		if err != nil {
			return
		}
	}
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
