package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	compileCommand := &cobra.Command{
		Use:   "compile",
		Short: "Compile a catalog into a graph to apply",
		RunE: func(cmd *cobra.Command, args []string) error {
			return compileCmd(cmd, args)
		},
	}
	rootCmd.AddCommand(compileCommand)
}

func compileCmd(c *cobra.Command, args []string) error {
	// Read file
	// Build graph, check for errors
	return errors.Errorf("Nothing to see here")
}
