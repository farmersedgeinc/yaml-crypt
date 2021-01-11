package actions

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/crypto"
	"github.com/farmersedgeinc/yaml-crypt/pkg/yaml"
	yamlv3 "gopkg.in/yaml.v3"
)

type nothing struct{}

func Decrypt(files []*File, plain bool, stdout bool, cache *cache.Cache, provider *crypto.Provider, threads int) (err error) {
	// read in files, populate the set of ciphertexts
	nodes := make([]yamlv3.Node, len(files))
	ciphertextSet := map[string]nothing{}
	for i, file := range files {
		nodes[i], err = yaml.ReadFile(file.EncryptedPath)
		if err != nil {
			return
		}
		err = addTaggedValuesToSet(&ciphertextSet, &nodes[i], yaml.EncryptedTag)
		if err != nil {
			return
		}
	}
	// fill in the cache with decryptions of all ciphertexts in the set
	err = decryptCiphertexts(&ciphertextSet, cache, provider, threads)
	if err != nil {
		return
	}
	for i, file := range files {
		// decrypt encrypted child nodes using now-loaded cache
		for node := range yaml.GetTaggedChildren(&nodes[i], yaml.EncryptedTag) {
			err = yaml.DecryptNode(node.YamlNode, cache, !plain)
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

func Encrypt(files []*File, cache *cache.Cache, provider *crypto.Provider, threads int) (err error) {
	// read in decrypted files, populate the set of plaintexts
	nodes := make([]yamlv3.Node, len(files))
	ciphertextSet := map[string]nothing{}
	plaintextSet := map[string]nothing{}
	for i, file := range files {
		nodes[i], err = yaml.ReadFile(file.DecryptedPath)
		if err != nil {
			return
		}
		err = addTaggedValuesToSet(&plaintextSet, &nodes[i], yaml.DecryptedTag)
		if err != nil {
			return
		}
		// if an encrypted version exists, load its encrypted values and add them to the ciphertext set, in order to later preload the cache with existing ciphertexts
		if exists(file.EncryptedPath) {
			var node yamlv3.Node
			node, err = yaml.ReadFile(file.EncryptedPath)
			if err != nil {
				return
			}
			err = addTaggedValuesToSet(&ciphertextSet, &node, yaml.EncryptedTag)
			if err != nil {
				return
			}
		}
	}
	// decrypt any encrypted values first, to pre-fill the cache with their existing versions
	err = decryptCiphertexts(&ciphertextSet, cache, provider, threads)
	if err != nil {
		return
	}
	// now we can encrypt any plaintexts that still don't have ciphertexts in the cache
	err = encryptPlaintexts(&plaintextSet, cache, provider, threads)
	if err != nil {
		return
	}

	for i, file := range files {
		// encrypt decrypted child nodes using now-loaded cache
		for node := range yaml.GetTaggedChildren(&nodes[i], yaml.DecryptedTag) {
			err = yaml.EncryptNode(node.YamlNode, cache)
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
			return err
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
				return "", err
			}
			err = cache.Add(plaintext, ciphertext)
			return string(ciphertext), err
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
