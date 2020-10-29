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

var Strings = []string{
	"1ZQheycERGjpTeXgJrzpjmxDxaixVJVcstgKyiwUshRx7AwZAsHtRwGVFgDtQlyVRiMMw618Hr4kbty66v1NN7acAnApEBAz9BfwfQ5kz87aGcaPSeck3F8obdyRPHmA",
	"xQr3HS4TmD87DPLMU17gUicZ",
	"lWBsiwTmlpA0H5k2nZjz64UM",
	"test",
	"{\"message\": \"weird, but could happen\"}",
	"\n % cat /dev/urandom | strings -n 16\n3 ,V\";u=)gg	H(>{d\nhV9+pprKG	>|=)IyO\nPvico@rXJ{L/g&g\n b83x)))n5GU+oI*a)\n!__dL`;5IX]/)ro1Pa4\nh;{a3U8\\fRYI}07Ivr]\nCy%JJY]kBMl!)tm	\nWo:Y	@1|5<FPFF	t }\n",
	"🤔🤔🤔",
}
