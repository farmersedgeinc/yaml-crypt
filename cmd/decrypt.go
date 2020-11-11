package cmd

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/spf13/cobra"
	"errors"
	"os"
)

var DecryptFlags struct {
	Stdout bool
	Plain bool
}

var DecryptCmd = &cobra.Command{
	Use:   "decrypt [file|directory]...",
	Short: "Decrypt the one or more files, creating a \"decrypted version\" that can be edited.",
	Long: "Decrypt the one or more files, creating a \"decrypted version\" that can be edited. Each arg can refer to either a file, in which case the file will be decrypted, or a directory, in which case all files under the directory will be decrypted. File args can refer to encrypted, decrypted, or plain files, existant or non-existant, as long as the correponding encrypted file exists. Supplying no args will decrypt all encrypted files in the repo.",
	Args:  func(cmd *cobra.Command, args []string) error {
		if DecryptFlags.Stdout && len(args) != 1 {
			return errors.New("requires exactly 1 arg when --stdout is set")
		}
		return nil
	},
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig(".")
		if err != nil { return err }
		if len(args) == 0 {
			args = []string{config.Root}
		}
		for _, arg := range args {
			var files []string
			if info, err := os.Stat(arg); !os.IsNotExist(err) && info.IsDir() {
				// if the arg is a dir, get all encrypted files in it
				files, err = config.AllEncryptedFiles(arg)
				if err != nil { return err }
			} else {
				// otherwise, just let actions.NewFile figure it out later
				files = []string{arg}
			}
			for _, file := range files {
				err = actions.Decrypt(actions.NewFile(file, &config), &config.Provider, DecryptFlags.Plain, DecryptFlags.Stdout, threads)
				if err != nil { return err }
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(DecryptCmd)
	DecryptCmd.Flags().BoolVarP(&DecryptFlags.Stdout, "stdout", "s", false, "print to stdout instead of saving to file")
	DecryptCmd.Flags().BoolVarP(&DecryptFlags.Plain, "plain", "p", false, "strip !secret tags from output yaml")
}
