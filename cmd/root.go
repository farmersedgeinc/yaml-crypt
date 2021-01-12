package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var threads uint
var progress bool

var rootCmd = &cobra.Command{
	Use:   "yaml-crypt",
	Short: "Encrypt secret values in your yaml files using a cloud-based encryption service.",
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		// Don't print usage for errors that happen after parsing has finished
		cmd.SilenceUsage = true
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().UintVarP(&threads, "threads", "t", 16, "number of crypto operations to run in parallel")
	rootCmd.PersistentFlags().BoolVarP(&progress, "progress", "", true, "show progress bar")
}
