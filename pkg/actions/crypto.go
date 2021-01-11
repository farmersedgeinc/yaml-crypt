package actions

import (
	"fmt"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/crypto"
	"github.com/farmersedgeinc/yaml-crypt/pkg/yaml"
	yamlv3 "gopkg.in/yaml.v3"
)

type nothing struct{}

func Decrypt(files []*File, plain bool, stdout bool, cache *cache.Cache, provider *crypto.Provider, threads int) error {
	// read in files, populate the set of ciphertexts
	var err error
	nodes := make([]yamlv3.Node, len(files))
	ciphertextSet := map[string]nothing{}
	for i, file := range files {
		nodes[i], err = yaml.ReadFile(file.EncryptedPath)
		if err != nil {
			return fmt.Errorf("Error reading yaml file %s: %w", file.DecryptedPath, err)
		}
		err = addTaggedValuesToSet(&ciphertextSet, &nodes[i], yaml.EncryptedTag)
		if err != nil {
			return fmt.Errorf("Error getting encrypted values from file %s: %w", file.EncryptedPath, err)
		}
	}
	// fill in the cache with decryptions of all ciphertexts in the set
	err = decryptCiphertexts(&ciphertextSet, cache, provider, threads)
	if err != nil {
		return fmt.Errorf("Error decrypting existing ciphertexts: %w", err)
	}
	for i, file := range files {
		// decrypt encrypted child nodes using now-loaded cache
		for node := range yaml.GetTaggedChildren(&nodes[i], yaml.EncryptedTag) {
			err = yaml.DecryptNode(node.YamlNode, cache, !plain)
			if err != nil {
				return fmt.Errorf("Error decrypting node %s using cache: %w", node.Path.String(), err)
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
			return fmt.Errorf("Error writing yaml file %s: %w", file.EncryptedPath, err)
		}
	}
	return err
}

func Encrypt(files []*File, cache *cache.Cache, provider *crypto.Provider, threads int) error {
	// read in decrypted files, populate the set of plaintexts
	var err error
	decryptedNodes := make([]yamlv3.Node, len(files))
	ciphertexts := make([]map[string]string, len(files))
	ciphertextSet := map[string]nothing{}
	plaintextSet := map[string]nothing{}
	for i, file := range files {
		decryptedNodes[i], err = yaml.ReadFile(file.DecryptedPath)
		if err != nil {
			return fmt.Errorf("Error reading yaml file %s: %w", file.DecryptedPath, err)
		}
		err = addTaggedValuesToSet(&plaintextSet, &decryptedNodes[i], yaml.DecryptedTag)
		if err != nil {
			return fmt.Errorf("Error getting decrypted values from file %s: %w", file.DecryptedPath, err)
		}
		// if an encrypted version exists, load its encrypted values and add them to the ciphertext set, in order to later preload the cache with existing ciphertexts
		if exists(file.EncryptedPath) {
			var node yamlv3.Node
			node, err = yaml.ReadFile(file.EncryptedPath)
			if err != nil {
				return fmt.Errorf("Error reading yaml file %s: %w", file.EncryptedPath, err)
			}
			ciphertexts[i], err = yaml.GetTaggedChildrenValues(&node, yaml.EncryptedTag)
			if err != nil {
				return fmt.Errorf("Error getting encrypted values from file %s: %w", file.EncryptedPath, err)
			}
			err = addTaggedValuesToSet(&ciphertextSet, &node, yaml.EncryptedTag)
			if err != nil {
				return fmt.Errorf("Error getting encrypted values from file %s: %w", file.EncryptedPath, err)
			}
		}
	}
	// decrypt any encrypted values first, to pre-fill the cache with their existing versions
	err = decryptCiphertexts(&ciphertextSet, cache, provider, threads)
	if err != nil {
		return fmt.Errorf("Error decrypting existing ciphertexts: %w", err)
	}
	// now we can encrypt any plaintexts that still don't have ciphertexts in the cache
	err = encryptPlaintexts(&plaintextSet, cache, provider, threads)
	if err != nil {
		return fmt.Errorf("Error encrypting plaintexts: %w", err)
	}

	for i, file := range files {
		// encrypt decrypted child nodes using now-loaded cache
		for node := range yaml.GetTaggedChildren(&decryptedNodes[i], yaml.DecryptedTag) {
			possibleCiphertext, _ := ciphertexts[i][node.Path.String()]
			err = yaml.EncryptNode(node.YamlNode, []byte(possibleCiphertext), cache)
			if err != nil {
				return fmt.Errorf("Error encrypting node %s using cache: %w", node.Path.String(), err)
			}
		}
		// write output
		err = yaml.SaveFile(file.EncryptedPath, decryptedNodes[i])
		if err != nil {
			return fmt.Errorf("Error writing yaml file %s: %w", file.EncryptedPath, err)
		}
	}
	return err
}

func addTaggedValuesToSet(set *map[string]nothing, node *yamlv3.Node, tag string) (err error) {
	values, err := yaml.GetTaggedChildrenValues(node, tag)
	if err != nil {
		return
	}
	for _, value := range values {
		(*set)[value] = nothing{}
	}
	return
}

func encryptPlaintexts(set *map[string]nothing, cache *cache.Cache, provider *crypto.Provider, threads int) error {
	var misses []string
	// get everything we can from the cache
	for plaintext := range *set {
		_, ok, err := cache.Encrypt(plaintext, []byte{})
		if err != nil {
			return fmt.Errorf("Error looking up plaintext in cache: %w", err)
		}
		if !ok {
			misses = append(misses, plaintext)
		}
	}
	// encrypt anything that missed the cache
	if len(misses) > 0 {
		_, err := parallelMap(misses, func(plaintext string) (string, error) {
			ciphertext, err := (*provider).Encrypt(plaintext)
			if err != nil {
				return "", fmt.Errorf("Error using provider to encrypt plaintext: %w", err)
			}
			err = cache.Add(plaintext, ciphertext)
			if err != nil {
				return "", fmt.Errorf("Error adding item to cache: %w", err)
			}
			return string(ciphertext), nil
		}, threads)
		if err != nil {
			return err
		}
	}
	return nil
}

func decryptCiphertexts(set *map[string]nothing, cache *cache.Cache, provider *crypto.Provider, threads int) error {
	var misses []string
	// attempt to "decrypt" with the cache
	for ciphertext := range *set {
		_, ok, err := cache.Decrypt([]byte(ciphertext))
		if err != nil {
			return err
		}
		if !ok {
			misses = append(misses, ciphertext)
		}
	}
	// decrypt anything that missed the cache, add it to the cache
	if len(misses) > 0 {
		_, err := parallelMap(misses, func(ciphertext string) (string, error) {
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
