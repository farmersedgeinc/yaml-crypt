package cmd

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/spf13/cobra"
)

var plainAllCmd = &cobra.Command{
	Use:   "plain-all",
	Short: "Decrypt all encrypted files, stripping all \"!encrypt\" tags.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig()
		if err != nil { return err }
		files, err := config.AllPlainFiles()
		if err != nil { return err }
		for _, file := range files {
			err = actions.Decrypt(actions.NewFile(file, &config), &config.Provider, false, false, threads)
			if err != nil { return err }
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(plainAllCmd)
}
