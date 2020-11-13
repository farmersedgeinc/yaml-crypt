package fixtures

import (
	"path/filepath"
	"io/ioutil"
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
)

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

func (f File) Compare(kind string) (bool, error) {
	eq, err := compareFiles(f.SrcPath(kind), f.TmpPath(kind))
	if !eq {
		fmt.Println(f.SrcPath(kind), f.TmpPath(kind))
	}
	return eq, err
}

func (f File) DiffDecrypted() ([]diffmatchpatch.Diff, error) {
	return diffFiles(f.SrcPath("original"), f.SrcPath("modified"))
}

func diffFiles(pathA, pathB string) ([]diffmatchpatch.Diff, error) {
	bytesA, err := ioutil.ReadFile(pathA)
	if err != nil {
		return []diffmatchpatch.Diff{}, err
	}
	bytesB, err := ioutil.ReadFile(pathB)
	if err != nil {
		return []diffmatchpatch.Diff{}, err
	}
	return DiffBytes(bytesA, bytesB), nil
}

func DiffBytes(a, b []byte) ([]diffmatchpatch.Diff) {
	d := diffmatchpatch.New()
	runesA, runesB, _ := d.DiffLinesToRunes(string(a), string(b))
	return d.DiffMainRunes(runesA, runesB, false)
}
