package actions

import (
	"fmt"
	"time"

	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/crypto"
	"github.com/farmersedgeinc/yaml-crypt/pkg/yaml"
	"github.com/schollz/progressbar/v3"
	yamlv3 "gopkg.in/yaml.v3"
)

type nothing struct{}

func Decrypt(files []*File, plain bool, stdout bool, cache *cache.Cache, provider *crypto.Provider, threads int, retries uint, timeout time.Duration, progress bool) error {
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
	err = decryptCiphertexts(&ciphertextSet, cache, provider, threads, retries, timeout, progress)
	if err != nil {
		return fmt.Errorf("Error decrypting existing ciphertexts: %w", err)
	}
	for i, file := range files {
		// decrypt encrypted child nodes using now-loaded cache
		for node := range yaml.GetTaggedChildren(&nodes[i], yaml.EncryptedTag) {
			err = yaml.DecryptNode(node.YamlNode, cache)
			if err != nil {
				return fmt.Errorf("Error decrypting node %s using cache: %w", node.Path.String(), err)
			}
		}
		// write modified root node out to file
		var err error
		if stdout {
			err = yaml.SaveFile("", nodes[i])
		} else if plain {
			yaml.StripTags(&nodes[i], yaml.DecryptedTag)
			err = yaml.SaveFile(file.PlainPath, nodes[i])
		} else {
			err = func() error {
				err := yaml.SaveFile(file.DecryptedPath, nodes[i])
				if err != nil {
					return err
				}
				if exists(file.PlainPath) {
					yaml.StripTags(&nodes[i], yaml.DecryptedTag)
					err = yaml.SaveFile(file.PlainPath, nodes[i])
				}
				return err
			}()
		}
		if err != nil {
			return fmt.Errorf("Error writing yaml file %s: %w", file.EncryptedPath, err)
		}
	}
	return err
}

func Encrypt(files []*File, cache *cache.Cache, provider *crypto.Provider, threads int, retries uint, timeout time.Duration, progress bool) error {
	// read in decrypted files, populate the set of plaintexts
	var err error
	decryptedNodes := make([]yamlv3.Node, len(files))
	ciphertextPathMaps := make([]map[string]string, len(files))
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
			ciphertextPathMaps[i], err = yaml.GetTaggedChildrenValues(&node, yaml.EncryptedTag)
			if err != nil {
				return fmt.Errorf("Error getting encrypted values from file %s: %w", file.EncryptedPath, err)
			}
			err = addTaggedValuesToSet(&ciphertextSet, &node, yaml.EncryptedTag)
			if err != nil {
				return fmt.Errorf("Error getting encrypted values from file %s: %w", file.EncryptedPath, err)
			}
		}
		// if a plain version exists, update it with the values we're encrypting.
		if exists(file.PlainPath) {
			clone := yaml.DeepCopyNode(&decryptedNodes[i])
			yaml.StripTags(clone, yaml.EncryptedTag)
			err = yaml.SaveFile(file.PlainPath, *clone)
		}
	}
	// decrypt any encrypted values first, to pre-fill the cache with their existing versions
	err = decryptCiphertexts(&ciphertextSet, cache, provider, threads, retries, timeout, progress)
	if err != nil {
		return fmt.Errorf("Error decrypting existing ciphertexts: %w", err)
	}
	// now we can encrypt any plaintexts that still don't have ciphertexts in the cache
	err = encryptPlaintexts(&plaintextSet, cache, provider, threads, retries, timeout, progress)
	if err != nil {
		return fmt.Errorf("Error encrypting plaintexts: %w", err)
	}

	for i, file := range files {
		// encrypt decrypted child nodes using now-loaded cache
		for node := range yaml.GetTaggedChildren(&decryptedNodes[i], yaml.DecryptedTag) {
			possibleCiphertext, _ := ciphertextPathMaps[i][node.Path.String()]
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

func encryptPlaintexts(set *map[string]nothing, cache *cache.Cache, provider *crypto.Provider, threads int, retries uint, timeout time.Duration, progress bool) error {
	plaintexts := make([]string, 0, len(*set))
	for k := range *set {
		plaintexts = append(plaintexts, k)
	}
	_, err := parallelMap(plaintexts, func(plaintext string) (string, error) {
		_, err := EncryptPlaintext(plaintext, cache, provider, retries, timeout)
		return "", err
	}, threads, progress)
	return err
}

func EncryptPlaintext(plaintext string, cache *cache.Cache, provider *crypto.Provider, retries uint, timeout time.Duration) ([]byte, error) {
	ciphertext, ok, err := cache.Encrypt(plaintext, []byte{})
	if err != nil {
		return []byte{}, fmt.Errorf("Error looking up plaintext in cache: %w", err)
	}
	if ok {
		return ciphertext, nil
	}
	ciphertext, err = (*provider).Encrypt(plaintext, retries, timeout)
	if err != nil {
		return []byte{}, fmt.Errorf("Error using provider to encrypt plaintext: %w", err)
	}
	err = cache.Add(plaintext, ciphertext)
	if err != nil {
		return []byte{}, fmt.Errorf("Error adding item to cache: %w", err)
	}
	return ciphertext, nil
}

func decryptCiphertexts(set *map[string]nothing, cache *cache.Cache, provider *crypto.Provider, threads int, retries uint, timeout time.Duration, progress bool) error {
	ciphertexts := make([]string, 0, len(*set))
	for k := range *set {
		ciphertexts = append(ciphertexts, k)
	}
	_, err := parallelMap(ciphertexts, func(ciphertext string) (string, error) {
		_, err := DecryptCiphertext([]byte(ciphertext), cache, provider, retries, timeout)
		return "", err
	}, threads, progress)
	return err
}

func DecryptCiphertext(ciphertext []byte, cache *cache.Cache, provider *crypto.Provider, retries uint, timeout time.Duration) (string, error) {
	plaintext, ok, err := cache.Decrypt(ciphertext)
	if err != nil {
		return "", fmt.Errorf("Error looking up ciphertext in cache: %w", err)
	}
	if ok {
		return plaintext, nil
	}
	plaintext, err = (*provider).Decrypt(ciphertext, retries, timeout)
	if err != nil {
		return "", fmt.Errorf("Error using provider to decrypt ciphertext: %w", err)
	}
	err = cache.Add(plaintext, ciphertext)
	if err != nil {
		return "", fmt.Errorf("Error adding item to cache: %w", err)
	}
	return plaintext, nil
}

func parallelMap(inputs []string, function func(string) (string, error), threads int, progress bool) (outputs map[string]string, err error) {
	inputChannel := make(chan string)
	outputChannel := make(chan mapResult)
	var bar *progressbar.ProgressBar
	if progress {
		bar = progressbar.NewOptions(
			len(inputs),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionShowCount(),
			progressbar.OptionSetPredictTime(false),
		)
	}
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
		if progress {
			bar.Add(1)
		}
	}
	if progress {
		bar.Finish()
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
