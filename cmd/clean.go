package cmd

import (
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/spf13/cobra"
	"os"
)

var cleanFlags struct {
	dir string
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Delete all decrypted files from the repo.",
	Long: "Clean all decrypted files from the repo. Warning: no attempt is made to \"shred\" the files as this doesn't work anymore on modern SSDs. Secrets may still be recoverable. Always use a machine with an encrypted disk for any sensitive data.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := os.Chdir(cleanFlags.dir)
		if err != nil { return err }
		config, err := config.LoadConfig()
		if err != nil { return err }
		decryptedFiles, err := config.AllDecryptedFiles()
		if err != nil { return err }
		plainFiles, err := config.AllPlainFiles()
		if err != nil { return err }
		for _, file := range append(decryptedFiles, plainFiles...) {
			err := os.Remove(file)
			if err != nil { return err }
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().StringVarP(&cleanFlags.dir, "dir", "d", ".", "path to start from when searching for the repo")
}
