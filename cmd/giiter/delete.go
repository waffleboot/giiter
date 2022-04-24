package main

import (
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete review branches",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		_featureBranch, err = git.FindFeatureBranch(cmd.Context(), _featureBranch)

		return
	},
	RunE: deleteReviewBranches,
}

func deleteReviewBranches(cmd *cobra.Command, args []string) error {
	reviewBranches, err := git.AllReviewBranches(cmd.Context(), _featureBranch)
	if err != nil {
		return err
	}

	for _, branch := range reviewBranches {
		if err := git.DeleteBranch(cmd.Context(), branch.BranchName()); err != nil {
			return err
		}
	}

	return nil
}
