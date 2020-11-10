package config

import (
	"gopkg.in/yaml.v3"
	"github.com/farmersedgeinc/yaml-crypt/pkg/crypto"
	"path/filepath"
	"strings"
	"os"
	"errors"
)

const ConfigFilename = ".yamlcrypt.yaml"

type SuffixesConfig struct {
	Encrypted string
	Decrypted string
	Plain string
}

func (c *SuffixesConfig) GitignoreSet() map[string]bool {
	return map[string]bool{
		"*." + c.Decrypted: true,
		"*." + c.Plain: true,
	}
}

func (c *SuffixesConfig) UnmarshalYAML(node *yaml.Node) error {
	tmp := map[string]string{}
	err := node.Decode(tmp)
	if err != nil {
		return err
	}
	if tmp["encrypted"] == "" {
		c.Encrypted = DefaultSuffixesConfig.Encrypted
	} else {
		c.Encrypted = strings.TrimPrefix(tmp["encrypted"], ".")
	}
	if tmp["decrypted"] == "" {
		c.Decrypted = DefaultSuffixesConfig.Decrypted
	} else {
		c.Decrypted = strings.TrimPrefix(tmp["decrypted"], ".")
	}
	if tmp["plain"] == "" {
		c.Plain = DefaultSuffixesConfig.Plain
	} else {
		c.Plain = strings.TrimPrefix(tmp["plain"], ".")
	}
	return nil
}

var DefaultSuffixesConfig = SuffixesConfig{
	Encrypted: "encrypted.yaml",
	Decrypted: "decrypted.yaml",
	Plain: "plain.yaml",
}

type Config struct {
	Provider crypto.Provider
	Suffixes SuffixesConfig
	Root string
}

func (c *Config) UnmarshalYAML(node *yaml.Node) error {
	type tmp struct {
		Provider string
		Config map[string] interface{}
		Suffixes SuffixesConfig
	}
	var t tmp
	err := node.Decode(&t)
	if err != nil {
		return err
	}

	provider, err := crypto.NewProvider(t.Provider, t.Config)
	if err != nil {
		return err
	}
	c.Provider = provider
	c.Suffixes = t.Suffixes
	return nil
}

func FindRepoRoot() (string, error) {
	path, err := findConfigFile()
	return filepath.Dir(path), err
}

func findConfigFile() (string, error) {
	for path, err := filepath.Abs("."); ; path = filepath.Dir(path) {
		if err != nil {
			return "", err
		}
		possiblePath := filepath.Join(path, ConfigFilename)
		_, err := os.Stat(possiblePath)
		if !os.IsNotExist(err) {
			return possiblePath, nil
		}
		if path == "/" {
			return "", errors.New("No config file found")
		}
	}
}

func LoadConfig() (Config, error) {
	var c Config
	path, err := findConfigFile()
	if err != nil {
		return c, err
	}
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return c, err
	}
	err = yaml.NewDecoder(f).Decode(&c)
	c.Root = filepath.Dir(path)
	return c, err
}

func (c *Config) allFiles(dir string, suffix string) ([]string, error) {
	var out []string
	err := filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) error {
			if (info == nil || !info.IsDir()) && strings.HasSuffix(path, suffix) {
				out = append(out, path)
			}
			if !os.IsNotExist(err) {
				return err
			}
			return nil
		},
	)
	return out, err
}

func(c *Config) AllEncryptedFiles(dir string) ([]string, error) {
	return c.allFiles(dir, c.Suffixes.Encrypted)
}

func(c *Config) AllDecryptedFiles(dir string) ([]string, error) {
	return c.allFiles(dir, c.Suffixes.Decrypted)
}

func(c *Config) AllPlainFiles(dir string) ([]string, error) {
	return c.allFiles(dir, c.Suffixes.Plain)
}
