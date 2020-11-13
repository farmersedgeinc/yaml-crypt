package cmd

import (
	"fmt"
	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/farmersedgeinc/yaml-crypt/pkg/crypto"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strconv"
)

var initFlags struct {
	provider string
	dir      string
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a yaml-crypt repo",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := os.Chdir(initFlags.dir)
		if err != nil {
			return err
		}
		path, err := config.FindRepoRoot(".")
		if err == nil {
			return fmt.Errorf("Repo already exists at %s", strconv.Quote(path))
		}
		providerConfig, ok := crypto.BlankConfigs[initFlags.provider]
		if !ok {
			return fmt.Errorf("Invalid provider name %s", strconv.Quote(initFlags.provider))
		}
		content := map[string]interface{}{
			"provider": initFlags.provider,
			"config":   providerConfig,
			"suffixes": config.DefaultSuffixesConfig,
		}
		out, err := yaml.Marshal(content)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(config.ConfigFilename, out, 0644)
		if err != nil {
			return err
		}
		config, err := config.LoadConfig(".")
		if err != nil {
			return err
		}
		err = actions.UpdateGitignore(&config)
		return err
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&initFlags.provider, "provider", "p", "", "name of the provider to use")
	initCmd.MarkFlagRequired("provider")
	initCmd.Flags().StringVarP(&initFlags.dir, "dir", "d", ".", "path to the root of the repo")
}
