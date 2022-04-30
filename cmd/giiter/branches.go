package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

type branchesCommand struct {
	featureBranch string
}

func makeBranchesCommand() *cobra.Command {
	var c branchesCommand

	cmd := &cobra.Command{
		Use:     "branches",
		Short:   "show all review branches",
		Aliases: []string{"b"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			c.featureBranch, err = git.FindFeatureBranch(cmd.Context(), c.featureBranch)

			return
		},
		RunE: c.run,
	}
	cmd.Flags().StringVarP(&c.featureBranch, "feature", "f", "", "feature branch")

	return cmd
}

func (c *branchesCommand) run(cmd *cobra.Command, args []string) error {
	reviewBranches, err := git.AllReviewBranches(cmd.Context(), c.featureBranch)
	if err != nil {
		return err
	}

	for _, branch := range reviewBranches {
		fmt.Printf("%s\n", branch.BranchName())
	}

	return nil
}
