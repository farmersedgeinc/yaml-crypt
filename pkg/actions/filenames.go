package actions

import (
	"errors"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	EncryptedPath string
	DecryptedPath string
	PlainPath     string
}

func NewFile(path string, config *config.Config) (File, error) {
	path, err := barePath(path, config)
	return File{
		EncryptedPath: path + config.Suffixes.Encrypted,
		DecryptedPath: path + config.Suffixes.Decrypted,
		PlainPath:     path + config.Suffixes.Plain,
	}, err
}

func barePath(path string, config *config.Config) (string, error) {
	dir := filepath.Dir(path)
	name := filepath.Base(path)
	length := -1
	for _, suffix := range []string{config.Suffixes.Encrypted, config.Suffixes.Decrypted, config.Suffixes.Plain} {
		if strings.HasSuffix(name, suffix) {
			length = len(suffix)
		}
	}
	if length == -1 {
		return "", errors.New("Filename does not end with any of the configured suffixes")
	}
	return filepath.Join(dir, name[:len(name)-length]), nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
