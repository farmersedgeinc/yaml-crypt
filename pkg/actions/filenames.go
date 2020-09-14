package actions

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"path/filepath"
	"strings"
	"errors"
	"os"
)

type File struct {
	Path string
	config *config.SuffixesConfig
}

func NewFile(path string, config *config.Config) *File {
	return &File{path, &config.Suffixes}
}

func (f *File)EncryptedPath() (string, error) {
	path, err := f.BarePath()
	if err != nil {
		return path, err
	}
	return path + f.config.Encrypted, nil
}

func (f *File)DecryptedPath() (string, error) {
	path, err := f.BarePath()
	if err != nil {
		return path, err
	}
	return path + f.config.Decrypted, nil
}

func (f *File)PlainPath() (string, error) {
	path, err := f.BarePath()
	if err != nil {
		return path, err
	}
	return path + f.config.Plain, nil
}

func (f *File) BarePath() (string, error) {
	dir := filepath.Dir(f.Path)
	name := filepath.Base(f.Path)
	length := -1
	for _, suffix := range []string{f.config.Encrypted, f.config.Decrypted, f.config.Plain} {
		if strings.HasSuffix(name, suffix) {
			length = len(suffix)
		}
	}
	if length == -1 {
		return "", errors.New("Filename does not end with any of the configured suffixes")
	}
	return filepath.Join(dir, name[:len(name)-length]), nil
}

func (f *File) AllPaths() (string, string, string, error) {
	encrypted, err := f.EncryptedPath()
	if err != nil { return "", "", "", err }
	decrypted, err := f.DecryptedPath()
	if err != nil { return "", "", "", err }
	plain, err := f.PlainPath()
	return encrypted, decrypted, plain, err
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
