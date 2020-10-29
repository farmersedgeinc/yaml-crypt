package actions

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/fixtures"
	"testing"
)
func TestFiles(t *testing.T) {
	for _, repo := range fixtures.Repos {
		config, err := config.LoadConfig(repo.Dir())
		if err != nil && repo.Exists {
			t.Errorf("Repo %s exists but loading gave error: %s", repo.Dir(), err.Error())
		} else if err == nil && !repo.Exists {
			t.Errorf("Repo %s does not exist but loading gave no error", repo.Dir())
		}
		// we can only plausibly run tests for fixtures with a valid config
		if err == nil {
			for _, file := range repo.Files {
				f := NewFile(file.Path(repo), &config)

				_, err := f.EncryptedPath()
				if err != nil && file.ValidName {
					t.Errorf("EncryptedPath() on Valid file %s raised error: %s", file.Path(repo), err.Error())
				} else if err == nil && !file.ValidName {
					t.Errorf("EncryptedPath() on Invalid file %s raised no error", file.Path(repo))
				}

				_, err = f.DecryptedPath()
				if err != nil && file.ValidName {
					t.Errorf("DecryptedPath() on Valid file %s raised error: %s", file.Path(repo), err.Error())
				} else if err == nil && !file.ValidName {
					t.Errorf("DecryptedPath() on Invalid file %s raised no error", file.Path(repo))
				}

				_, err = f.PlainPath()
				if err != nil && file.ValidName {
					t.Errorf("PlainPath() on Valid file %s raised error: %s", file.Path(repo), err.Error())
				} else if err == nil && !file.ValidName {
					t.Errorf("PlainPath() on Invalid file %s raised no error", file.Path(repo))
				}
				_, _, _, err = f.AllPaths()
				if err != nil && file.ValidName {
					t.Errorf("AllPaths() on Valid file %s raised error: %s", file.Path(repo), err.Error())
				} else if err == nil && !file.ValidName {
					t.Errorf("AllPaths() on Invalid file %s raised no error", file.Path(repo))
				}
			}
		}
	}
}
