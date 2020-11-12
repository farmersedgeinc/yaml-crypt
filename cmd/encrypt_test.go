package cmd

import (
	"testing"
	"github.com/farmersedgeinc/yaml-crypt/pkg/fixtures"
	"io/ioutil"
	"bytes"
)

func TestEncrypt(t *testing.T) {
	repos, err := fixtures.Repos()
	if err != nil {
		t.Fatal(err)
	}
	for _, repo := range repos {
		DecryptFlags.Plain = false
		err := repo.Setup()
		defer repo.Destroy()
		if err != nil {
			t.Fatal(err)
		}
		err = repo.Checkout("original")
		if err != nil {
			t.Fatal(err)
		}

		// initial encrypt
		err = EncryptCmd.RunE(nil, []string{})
		if err != nil {
			t.Fatal(err)
		}

		// store initial encrypted outputs
		originalEncryptions := make([][]byte, len(repo.Files))
		for i, file := range repo.Files {
			originalEncryptions[i], err = ioutil.ReadFile(file.TmpPath(repo.Provider))
			if err != nil {
				t.Fatal(err)
			}
		}

		// second encrypt
		err = EncryptCmd.RunE(nil, []string{})
		if err != nil {
			t.Fatal(err)
		}

		// compare. outputs should be identical
		for i, file := range repo.Files {
			data, err := ioutil.ReadFile(file.TmpPath(repo.Provider))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, originalEncryptions[i]) {
				t.Errorf("Encrypted file %s in repo %s changed despite no changes to decrypted file!", file.TmpPath(repo.Provider), repo)
			}
		}

		// checkout modified file, encrypt that
		err = repo.Checkout("modified")
		err = EncryptCmd.RunE(nil, []string{})
		if err != nil {
			t.Fatal(err)
		}

		// compare. outputs should be different
		for i, file := range repo.Files {
			data, err := ioutil.ReadFile(file.TmpPath(repo.Provider))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(data, originalEncryptions[i]) {
				t.Errorf("Encrypted file %s in repo %s did not change despite changes to decrypted file!", file.TmpPath(repo.Provider), repo)
			}
		}
	}
}
