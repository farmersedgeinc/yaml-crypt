package cmd

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/spf13/cobra"
)

var plainFlags struct {
	stdout bool
}

var plainCmd = &cobra.Command{
	Use:   "plain",
	Short: "Decrypt an encrypted file, stripping all \"!encrypt\" tags.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig()
		if err != nil { return err }
		files, err := config.AllEncryptedFiles()
		if err != nil { return err }
		for _, file := range files {
			err = actions.Decrypt(actions.NewFile(file, &config), &config.Provider, true, plainFlags.stdout, threads)
			if err != nil { return err }
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(plainCmd)
	plainCmd.Flags().BoolVarP(&plainFlags.stdout, "stdout", "s", false, "print to stdout instead of saving to file")
}
