package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

var branchesCmd = &cobra.Command{
	Use:     "branches",
	Short:   "show all branches",
	Aliases: []string{"b"},
	RunE:    showAllBranches,
}

func showAllBranches(cmd *cobra.Command, args []string) error {
	branches, err := git.AllBranches(cmd.Context())
	if err != nil {
		return err
	}

	for i := range branches {
		fmt.Printf("%s\n", branches[i].Name)
	}

	return nil
}
