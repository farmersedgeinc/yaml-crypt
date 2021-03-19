package cmd

import (
	"bufio"
	"encoding/base64"
	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

var decryptValueFlags struct {
}

var decryptValueCmd = &cobra.Command{
	Use:                   "decrypt-value",
	Short:                 "Read in an encrypted value from STDIN and print an decrypted representation to STDOUT",
	Long:                  "Read in an encrypted value from STDIN and print an decrypted representation to STDOUT. By default it will read in and decrypt a single line.",
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return DecryptValue(os.Stdin, os.Stdout)
	},
}

func DecryptValue(stdin io.Reader, stdout io.Writer) error {
	reader := bufio.NewReader(stdin)
	encodedCiphertext, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	encodedCiphertext = strings.TrimSpace(encodedCiphertext)
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return err
	}
	var plaintext string
	func() error {
		config, err := config.LoadConfig(".")
		if err != nil {
			return err
		}
		cache, err := cache.Setup(config)
		if err != nil {
			return err
		}
		defer cache.Close()
		plaintext, err = actions.DecryptCiphertext(ciphertext, &cache, &config.Provider)
		return err
	}()
	if err != nil {
		return err
	}
	io.WriteString(stdout, plaintext+"\n")
	return nil

}

func init() {
	rootCmd.AddCommand(decryptValueCmd)
}
