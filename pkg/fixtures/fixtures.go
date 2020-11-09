package fixtures

import (
	"path/filepath"
	"runtime"
	"io/ioutil"
	"strings"
	"fmt"
	"os"
	"errors"
	"io"
	"gopkg.in/yaml.v3"
	"golang.org/x/oauth2/google"
	"context"
)

func TestDataDir() (string, error) {
	// get path to this source file
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("Unable to find test dir: runtime.Caller failed")
	}
	root, err := filepath.Abs(here)
	if err != nil {
		return "", err
	}
	for i := 0; i < 3; i++ {
		root = filepath.Dir(root)
	}
	return filepath.Join(root, "testdata"), nil
}

type Repo struct {
	ConfigFile string
	Provider string
	Note string
	Files []File
	TmpDir string
	Suffixes map[string]string
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

func cp(src, dst string) error {
	// open source
	srcHandle, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcHandle.Close()
	// open destination
	dstHandle, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstHandle.Close()

	// copy
	_, err = io.Copy(dstHandle, srcHandle)
	if err != nil {
		return err
	}
	return dstHandle.Sync()
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
				Provider: parts[0],
				Suffixes: suffixes,
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
	return err
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

func (r Repo) Destroy() error {
	return os.RemoveAll(r.TmpDir)
}

type File struct {
	Name string
	SrcDir string
	Repo *Repo
}

func Files(repo *Repo) ([]File, error) {
	out := []File{}
	testDir, err := TestDataDir()
	if err != nil {
		return out, err
	}
	files, err := ioutil.ReadDir(filepath.Join(testDir, "files"))
	if err != nil {
		return out, err
	}
	for _, f := range files {
		if f.IsDir() {
			out = append(out, File{
				Name: f.Name(),
				SrcDir: filepath.Join(testDir, "files", f.Name()),
				Repo: repo,
			})
		}
	}
	return out, nil
}

func (f File) Checkout(kind string) error {
	return cp(f.SrcPath(kind), f.TmpPath(kind))
}

func (f File) SrcPath(kind string) string {
	return filepath.Join(f.SrcDir, kind + ".yaml")
}

func (f File) TmpPath(kind string) string {
	var name string
	if kind == "original" || kind == "modified" {
		name = f.Name + "." + f.Repo.Suffixes["decrypted"]
	} else if kind == "plain" {
		name  = f.Name + "." + f.Repo.Suffixes["plain"]
	} else {
		name  = f.Name + "." + f.Repo.Suffixes["encrypted"]
	}
	return filepath.Join(f.Repo.TmpDir, name)
}

var Strings = []string{
	"1ZQheycERGjpTeXgJrzpjmxDxaixVJVcstgKyiwUshRx7AwZAsHtRwGVFgDtQlyVRiMMw618Hr4kbty66v1NN7acAnApEBAz9BfwfQ5kz87aGcaPSeck3F8obdyRPHmA",
	"xQr3HS4TmD87DPLMU17gUicZ",
	"lWBsiwTmlpA0H5k2nZjz64UM",
	"test",
	"{\"message\": \"weird, but could happen\"}",
	"\n % cat /dev/urandom | strings -n 16\n3 ,V\";u=)gg	H(>{d\nhV9+pprKG	>|=)IyO\nPvico@rXJ{L/g&g\n b83x)))n5GU+oI*a)\n!__dL`;5IX]/)ro1Pa4\nh;{a3U8\\fRYI}07Ivr]\nCy%JJY]kBMl!)tm	\nWo:Y	@1|5<FPFF	t }\n",
	"🤔🤔🤔",
}
