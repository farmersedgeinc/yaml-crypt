package cmd

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/spf13/cobra"
)

var encryptAllFlags struct {
	dir string
}

var encryptAllCmd = &cobra.Command{
	Use:   "encrypt-all",
	Short: "Encrypt all the decrypted files in the repo, replacing the contents of the encrypted files.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig()
		if err != nil { return err }
		files, err := config.AllDecryptedFiles()
		if err != nil { return err }
		for _, file := range files {
			err = actions.Encrypt(actions.NewFile(file, &config), &config.Provider, threads)
			if err != nil { return err }
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(encryptAllCmd)
	encryptAllCmd.Flags().StringVarP(&encryptAllFlags.dir, "dir", "d", ".", "path to start from when searching for the repo")
}
