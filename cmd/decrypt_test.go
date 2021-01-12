package cmd

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/fixtures"
	"testing"
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
			eq, err := repo.Compare("original")
			if err != nil {
				t.Error(err.Error())
			}
			if !eq {
				t.Errorf("Decrypted files in repo %s are incorrect", repo)
			}
		}
		DecryptFlags.Plain = true
		err = DecryptCmd.RunE(nil, []string{})
		if err != nil {
			t.Error(err.Error())
		}
		eq, err := repo.Compare("plain")
		if err != nil {
			t.Error(err.Error())
		}
		if !eq {
			t.Errorf("Plain files in repo %s are incorrect", repo)
		}
	}
}
