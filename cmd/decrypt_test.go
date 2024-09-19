package cmd

import (
	"testing"

	"github.com/farmersedgeinc/yaml-crypt/pkg/fixtures"
)

func TestDecrypt(t *testing.T) {
	progress = false
	repos, err := fixtures.Repos()
	if err != nil {
		t.Fatal(err)
	}
	for _, repo := range repos {
		DecryptFlags.Plain = false
		err := repo.Setup()
		//defer repo.Destroy()
		if err != nil {
			t.Fatal(err)
		}
		err = repo.Checkout(repo.Provider)
		if err != nil {
			t.Fatal(err)
		}
		// test it twice since decrypting to an existing file is a different code path
		for i := 0; i < 2; i++ {
			err = DecryptCmd.RunE(nil, []string{})
			if err != nil {
				t.Error(err.Error())
			}
			files, err := repo.Compare("original")
			if err != nil {
				t.Error(err.Error())
			}
			if len(files) > 0 {
				names := []string{}
				for _, file := range files {
					names = append(names, file.Name)
				}
				t.Errorf("Decrypted files %v in repo %q are incorrect", names, repo)
			}
		}
		DecryptFlags.Plain = true
		err = DecryptCmd.RunE(nil, []string{})
		if err != nil {
			t.Error(err.Error())
		}
		files, err := repo.Compare("plain")
		if err != nil {
			t.Error(err.Error())
		}
		if len(files) > 0 {
			names := []string{}
			for _, file := range files {
				names = append(names, file.Name)
			}
			t.Errorf("Plain files %v in repo %q are incorrect", names, repo)
		}
	}
}
