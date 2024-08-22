package cmd

import (
	"os"

	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/spf13/cobra"
)

var EncryptCmd = &cobra.Command{
	Use:                   "encrypt [file|directory]...",
	Short:                 "Encrypt one or more decrypted files in the repo, replacing the contents of the encrypted files.",
	Long:                  "Encrypt one or more decrypted files in the repo, replacing the contents of the corresponding encrypted files. Each arg can refer to either a file, in which case the file will be encrypted, or a directory, in which case all files under the directory will be encrypted. File args can refer to encrypted, decrypted, or plain files, existant or non-existant, as long as the correponding decrypted file exists. Supplying no args will encrypt all decrypted files in the repo.",
	Args:                  cobra.ArbitraryArgs,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig(".")
		if err != nil {
			return err
		}
		cache, err := cache.Setup(config, disableCache)
		if err != nil {
			return err
		}
		defer cache.Close()
		if len(args) == 0 {
			args = []string{config.Root}
		}
		files := make([]*actions.File, 0, len(args))
		for _, arg := range args {
			var paths []string
			if info, err := os.Stat(arg); !os.IsNotExist(err) && info.IsDir() {
				// if the arg is a dir, get all encrypted files in it
				paths, err = config.AllDecryptedFiles(arg)
				if err != nil {
					return err
				}
			} else {
				// otherwise, just let actions.NewFile figure it out later
				paths = []string{arg}
			}
			for _, path := range paths {
				file, err := actions.NewFile(path, &config)
				if err != nil {
					return err
				}
				files = append(files, &file)
			}
		}
		return actions.Encrypt(files, cache, &config.Provider, int(threads), retries, timeout, progress)
	},
}

func init() {
	rootCmd.AddCommand(EncryptCmd)
}
