package cmd

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/spf13/cobra"
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt <file...>",
	Short: "Encrypt one or more decrypted files in the repo, replacing the contents of the encrypted files.",
	Args:  cobra.MinimumNArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig()
		if err != nil { return err }
		for _, file := range args {
			err = actions.Encrypt(actions.NewFile(file, &config), &config.Provider, threads)
			if err != nil { return err }
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(encryptCmd)
}
