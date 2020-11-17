package cmd

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/spf13/cobra"
)

var gitignoreFlags struct {
	dir string
}

var gitignoreCmd = &cobra.Command{
	Use:   "update-gitignore",
	Short: "Update the .gitignore file for this repo.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig(gitignoreFlags.dir)
		if err != nil {
			return err
		}
		err = actions.UpdateGitignore(&config)
		return err
	},
}

func init() {
	rootCmd.AddCommand(gitignoreCmd)
	gitignoreCmd.Flags().StringVarP(&gitignoreFlags.dir, "dir", "d", ".", "path to the root of the repo")
}
