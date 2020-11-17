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
		defer repo.Destroy()
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
				_, err := NewFile(path, &config)
				if err != nil {
					t.Errorf("NewFile() on File %s in Repo %s raised error: %s", path, repo, err.Error())
				}
			}
		}
	}
}
