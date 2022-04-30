package main

import (
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

type deleteCommand struct {
	featureBranch string
}

func makeDeleteCommand() *cobra.Command {
	var c deleteCommand

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete review branches",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			c.featureBranch, err = git.FindFeatureBranch(cmd.Context(), c.featureBranch)

			return
		},
		RunE: c.run,
	}
	cmd.Flags().StringVarP(&c.featureBranch, "feature", "f", "", "feature branch")

	return cmd
}

func (c *deleteCommand) run(cmd *cobra.Command, args []string) error {
	reviewBranches, err := git.AllReviewBranches(cmd.Context(), c.featureBranch)
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
