package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VersionString string = "v0.0.1"

func init() {
	versionCommand := &cobra.Command{
		Use:   "version",
		Short: "Displays jcfg version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return versionCmd(cmd, args)
		},
	}
	rootCmd.AddCommand(versionCommand)
}

func versionCmd(c *cobra.Command, args []string) error {
	// Read file
	// Build graph, check for errors
	fmt.Printf("jcfg version %s\n", VersionString)

	return nil
}
