package cmd

import (
	"bufio"
	"encoding/base64"
	"io"
	"os"
	"strings"

	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/spf13/cobra"
)

var decryptValueFlags struct {
	no_newline bool
}

var decryptValueCmd = &cobra.Command{
	Use:                   "decrypt-value",
	Short:                 "Read in an encrypted value from STDIN and print an decrypted representation to STDOUT",
	Long:                  "Read in an encrypted value from STDIN and print an decrypted representation to STDOUT. By default it will read in and decrypt a single line.",
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return DecryptValue(os.Stdin, os.Stdout, decryptValueFlags.no_newline)
	},
}

func DecryptValue(stdin io.Reader, stdout io.Writer, no_newline bool) error {
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
	err = func() error {
		config, err := config.LoadConfig(".")
		if err != nil {
			return err
		}
		cache, err := cache.Setup(config)
		if err != nil {
			return err
		}
		defer cache.Close()
		plaintext, err = actions.DecryptCiphertext(ciphertext, &cache, &config.Provider, retries, timeout)
		return err
	}()
	if err != nil {
		return err
	}
	io.WriteString(stdout, plaintext)
	if !no_newline {
		io.WriteString(stdout, "\n")
	}
	return nil

}

func init() {
	rootCmd.AddCommand(decryptValueCmd)
	decryptValueCmd.Flags().BoolVarP(&decryptValueFlags.no_newline, "no-newline", "n", false, "do not print a trailing newline.")
}
