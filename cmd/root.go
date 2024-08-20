package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
)

var threads uint
var progress bool
var disableCache bool
var retries uint
var timeout time.Duration

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
	rootCmd.PersistentFlags().BoolVarP(&disableCache, "no-cache", "C", false, "disable caching plaintexts to disk")
	rootCmd.PersistentFlags().UintVarP(&retries, "retries", "r", 5, "number of retries for failed crypto service operations")
	rootCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "", 10*time.Second, "timeout for crypto service operations")
}
