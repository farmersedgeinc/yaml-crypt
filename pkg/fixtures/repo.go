package fixtures

import (
	"context"
	"fmt"
	"golang.org/x/oauth2/google"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Repo struct {
	ConfigFile string
	Provider   string
	Note       string
	Files      []File
	TmpDir     string
	Suffixes   map[string]string
	OldCwd     string
}

func (r Repo) String() string {
	if r.Note == "" {
		return r.Provider
	} else {
		return fmt.Sprintf("%s.%s", r.Provider, r.Note)
	}
}

func SuffixesConfig(file string) (map[string]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return map[string]string{}, err
	}
	defer f.Close()

	type tmp struct {
		Suffixes map[string]interface{}
	}
	var t tmp
	err = yaml.NewDecoder(f).Decode(&t)
	if err != nil {
		return map[string]string{}, err
	}
	out := map[string]string{}
	for _, key := range []string{"encrypted", "decrypted", "plain"} {
		suffix, ok := t.Suffixes[key]
		if ok {
			out[key], ok = suffix.(string)
			if !ok {
				return out, fmt.Errorf("%s config value must be type string", key)
			}
		} else {
			out[key] = key + ".yaml"
		}
	}
	return out, nil
}

func Repos() ([]Repo, error) {
	out := []Repo{}
	testDir, err := TestDataDir()
	if err != nil {
		return out, err
	}
	files, err := ioutil.ReadDir(filepath.Join(testDir, "repos"))
	if err != nil {
		return out, err
	}
	for _, f := range files {
		if !f.IsDir() {
			parts := strings.Split(f.Name(), ".")
			if l := len(parts); l < 2 || l > 3 || parts[len(parts)-1] != "yaml" {
				return out, fmt.Errorf("Invalid repo fixture file %s", f.Name())
			}
			suffixes, err := SuffixesConfig(filepath.Join(testDir, "repos", f.Name()))
			if err != nil {
				return out, err
			}
			repo := Repo{
				ConfigFile: f.Name(),
				Provider:   parts[0],
				Suffixes:   suffixes,
			}
			repo.Files, err = Files(&repo)
			if err != nil {
				return out, err
			}
			if len(parts) == 3 {
				repo.Note = parts[1]
			}
			out = append(out, repo)
		}
	}
	return out, nil
}

func (r *Repo) Skip() bool {
	if r.Provider == "google" {
		_, err := google.FindDefaultCredentials(context.Background())
		return err != nil
	}
	return false
}

func (r *Repo) Setup() error {
	testDir, err := TestDataDir()
	if err != nil {
		return err
	}
	r.TmpDir, err = ioutil.TempDir(os.TempDir(), "yamlcrypt-test-repo-*")
	if err != nil {
		return err
	}
	err = cp(filepath.Join(testDir, "repos", r.ConfigFile), filepath.Join(r.TmpDir, ".yamlcrypt.yaml"))
	if err != nil {
		return err
	}
	r.OldCwd, err = os.Getwd()
	if err != nil {
		return err
	}
	return os.Chdir(r.TmpDir)
}

func (r Repo) Destroy() error {
	err := os.Chdir(r.OldCwd)
	if err != nil {
		return err
	}
	return os.RemoveAll(r.TmpDir)
}

func (r Repo) Checkout(kind string) error {
	for _, file := range r.Files {
		err := file.Checkout(kind)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r Repo) Compare(kind string) (bool, error) {
	for _, file := range r.Files {
		eq, err := file.Compare(kind)
		if err != nil {
			return false, err
		}
		if !eq {
			return false, nil
		}
	}
	return true, nil
}
