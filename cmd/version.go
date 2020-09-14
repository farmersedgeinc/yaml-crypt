package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
)

var version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "get the version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
