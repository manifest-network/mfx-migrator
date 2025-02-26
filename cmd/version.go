package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the application version. It can be set during build time using:
	// go build -ldflags "-X cmd.Version=x.y.z"
	Version = "dev"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
