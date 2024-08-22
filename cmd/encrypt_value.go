package cmd

import (
	"bufio"
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/spf13/cobra"
)

var encryptValueFlags struct {
	multiline bool
}

var encryptValueCmd = &cobra.Command{
	Use:                   "encrypt-value",
	Short:                 "Read in a decrypted value from STDIN and print an encrypted representation to STDOUT",
	Long:                  "Read in a decrypted value from STDIN and print an encrypted representation to STDOUT. By default it will read in and encrypt a single line.",
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return EncryptValue(os.Stdin, os.Stdout, encryptValueFlags.multiline)
	},
}

func EncryptValue(stdin io.Reader, stdout io.Writer, multiline bool) error {
	var plaintext string
	var err error
	if multiline {
		rawPlaintext, err := ioutil.ReadAll(stdin)
		if err != nil {
			return err
		}
		plaintext = string(rawPlaintext)
	} else {
		reader := bufio.NewReader(stdin)
		plaintext, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
		plaintext = strings.Trim(plaintext, "\n")
	}
	var ciphertext []byte
	err = func() error {
		config, err := config.LoadConfig(".")
		if err != nil {
			return err
		}
		cache, err := cache.Setup(config, disableCache)
		if err != nil {
			return err
		}
		defer cache.Close()
		ciphertext, err = actions.EncryptPlaintext(string(plaintext), cache, &config.Provider, retries, timeout)
		return err
	}()
	if err != nil {
		return err
	}
	io.WriteString(stdout, base64.StdEncoding.EncodeToString(ciphertext)+"\n")
	return nil
}

func init() {
	rootCmd.AddCommand(encryptValueCmd)
	encryptValueCmd.Flags().BoolVarP(&encryptValueFlags.multiline, "multi-line", "m", false, "Read multiple lines of input, stopping only at EOF.")
}
