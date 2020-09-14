package cmd

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/spf13/cobra"
)

var decryptAllFlags struct {
	dir string
}

var decryptAllCmd = &cobra.Command{
	Use:   "decrypt-all",
	Short: "Decrypt all the files in the repo, creating a \"decrypted versions\" that can be edited.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig()
		if err != nil { return err }
		files, err := config.AllDecryptedFiles()
		if err != nil { return err }
		for _, file := range files {
			err = actions.Decrypt(actions.NewFile(file, &config), &config.Provider, false, false, threads)
			if err != nil { return err }
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(decryptAllCmd)
	decryptAllCmd.Flags().StringVarP(&decryptAllFlags.dir, "dir", "d", ".", "path to start from when searching for the repo")
}
