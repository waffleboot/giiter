package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var branchesCmd = &cobra.Command{
	Use:     "branches",
	Short:   "show all branches",
	Aliases: []string{"b"},
	RunE:   showAllBranches,
}

func showAllBranches(cmd *cobra.Command, args []string) error {
	branches, err := g.Branches(cmd.Context())
	if err != nil {
		return err
	}

	for i := range branches {
		fmt.Printf("%s\n", branches[i].Name)
	}

	return nil
}
