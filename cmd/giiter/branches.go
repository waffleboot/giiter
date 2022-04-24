package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

var branchesCmd = &cobra.Command{
	Use:     "branches",
	Short:   "show all review branches",
	Aliases: []string{"b"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		_featureBranch, err = git.FindFeatureBranch(cmd.Context(), _featureBranch)

		return
	},
	RunE: showAllBranches,
}

func showAllBranches(cmd *cobra.Command, args []string) error {
	reviewBranches, err := git.AllReviewBranches(cmd.Context(), _featureBranch)
	if err != nil {
		return err
	}

	for _, branch := range reviewBranches {
		fmt.Printf("%s\n", branch.BranchName())
	}

	return nil
}
