package actions

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/fixtures"
	"testing"
)

func TestFiles(t *testing.T) {
	repos, err := fixtures.Repos()
	if err != nil {
		t.Fatal(err)
	}
	for _, repo := range repos {
		err := repo.Setup()
		// defer repo.Destroy()
		if err != nil {
			t.Fatal(err)
		}
		config, err := config.LoadConfig(repo.TmpDir)
		if err != nil {
			t.Fatalf("Loading repo %s gave error: %s", repo, err.Error())
		}
		for _, file := range repo.Files {
			for _, kind := range []string{"original", "noop", "plain"} {
				err = repo.Checkout(kind)
				if err != nil {
					t.Fatal(err)
				}
				path := file.TmpPath(kind)
				f := NewFile(path, &config)

				_, err := f.EncryptedPath()
				if err != nil {
					t.Errorf("EncryptedPath() on File %s in Repo %s raised error: %s", path, repo, err.Error())
				}

				_, err = f.DecryptedPath()
				if err != nil {
					t.Errorf("DecryptedPath() on File %s in Repo %s raised error: %s", path, repo, err.Error())
				}

				_, err = f.PlainPath()
				if err != nil {
					t.Errorf("PlainPath() on File %s in Repo %s raised error: %s", path, repo, err.Error())
				}
				_, _, _, err = f.AllPaths()
				if err != nil {
					t.Errorf("AllPaths() on File %s in Repo %s raised error: %s", path, repo, err.Error())
				}
			}
		}
	}
}
