package fixtures

import (
	"path/filepath"
)

var Repos = []Repo{
	Repo{
		dir: "repo0",
		ValidConfig: true,
		Exists: true,
		Files: []File{
			File{"example0.decrypted.yaml", true, true, true},
			File{"example1.encrypted.yaml", true, true, true},
			File{"example2.weirdsuffix.yaml", true, false, true},
			File{"example3_invalidyaml.decrypted.yaml", true, true, false},
			File{"example3_nonexistent.decrypted.yaml", false, true, false},
		},
	},
	Repo{
		dir: "repo1",
		ValidConfig: true,
		Exists: true,
		Files: []File{},
	},
	Repo{
		dir: "badconfig",
		ValidConfig: false,
		Exists: true,
		Files: []File{},
	},
	Repo{
		dir: "nonexistant",
		ValidConfig: false,
		Exists: false,
		Files: []File{
			File{"bad", false, false, false},
		},
	},
}

type Repo struct {
	dir string
	ValidConfig bool
	Exists bool
	Files []File
}

func (r Repo) Dir() string {
	return filepath.Join("../../test", r.dir)
}

type File struct {
	Name string
	Exists bool
	ValidName bool
	ValidYaml bool
}

func (f File) Path(r Repo) string {
	return filepath.Join(r.Dir(), f.Name)
}
