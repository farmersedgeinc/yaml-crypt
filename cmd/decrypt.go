package cmd

import (
	"errors"
	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/spf13/cobra"
	"os"
)

var DecryptFlags struct {
	Stdout bool
	Plain  bool
}

var DecryptCmd = &cobra.Command{
	Use:   "decrypt [file|directory]...",
	Short: "Decrypt the one or more files, creating a \"decrypted version\" that can be edited.",
	Long:  "Decrypt the one or more files, creating a \"decrypted version\" that can be edited. Each arg can refer to either a file, in which case the file will be decrypted, or a directory, in which case all files under the directory will be decrypted. File args can refer to encrypted, decrypted, or plain files, existant or non-existant, as long as the correponding encrypted file exists. Supplying no args will decrypt all encrypted files in the repo.",
	Args: func(cmd *cobra.Command, args []string) error {
		if DecryptFlags.Stdout && len(args) != 1 {
			return errors.New("requires exactly 1 arg when --stdout is set")
		}
		return nil
	},
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig(".")
		if err != nil {
			return err
		}
		cache, err := cache.Setup(config)
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
				paths, err = config.AllEncryptedFiles(arg)
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
		return actions.Decrypt(files, DecryptFlags.Plain, DecryptFlags.Stdout, &cache, &config.Provider, int(threads))
	},
}

func init() {
	rootCmd.AddCommand(DecryptCmd)
	DecryptCmd.Flags().BoolVarP(&DecryptFlags.Stdout, "stdout", "s", false, "print to stdout instead of saving to file")
	DecryptCmd.Flags().BoolVarP(&DecryptFlags.Plain, "plain", "p", false, "strip !secret tags from output yaml")
}
