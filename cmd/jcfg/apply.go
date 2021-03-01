package main

import (
	"context"

	"example.com/jcfg/pkg/catalog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	applyCommand := &cobra.Command{
		Use:   "apply GraphFile",
		Short: "Apply a json file",
		Long:  `Read in a compiled catalog/graph and enforce that state on the system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return applyCmd(cmd, args)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				log.Fatalf(errors.New("no json file specified on command line").Error())
			}
			return nil
		},
	}
	rootCmd.AddCommand(applyCommand)
}

func applyCmd(c *cobra.Command, args []string) error {
	setupLogger()
	// Read file
	graphFile := args[0]
	// Build graph
	loadedCatalog, err := catalog.NewCatalog(graphFile, log)
	if err != nil {
		return errors.Errorf("Error building catalog to apply: %s", err)
	}
	g := catalog.Graph{}
	if err := g.LoadCatalog(loadedCatalog, log); err != nil {
		return errors.Errorf("Error loading catalog into graph to apply: %s", err)
	}
	// Walk graph, applying resources
	ctx := context.Background()
	if err := g.Apply(ctx, log); err != nil {
		return errors.Errorf("Error applying catalog: %s", err)
	}
	return nil
}
