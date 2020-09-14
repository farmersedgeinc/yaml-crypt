package cmd

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/spf13/cobra"
)

var decryptFlags struct {
	stdout bool
}

var decryptCmd = &cobra.Command{
	Use:   "decrypt <file...>",
	Short: "Decrypt the one or more files, creating a \"decrypted version\" that can be edited.",
	Args:  cobra.MinimumNArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig()
		if err != nil { return err }
		for _, file := range args {
			err = actions.Decrypt(actions.NewFile(file, &config), &config.Provider, false, decryptFlags.stdout, threads)
			if err != nil { return err }
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)
	decryptCmd.Flags().BoolVarP(&decryptFlags.stdout, "stdout", "s", false, "print to stdout instead of saving to file")
}
