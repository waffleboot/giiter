package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/waffleboot/giiter/internal/git"
)

type branchesCommand struct {
	config *git.Config
}

func makeBranchesCommand(config *git.Config) *cobra.Command {
	var featureBranch string

	c := branchesCommand{
		config: config,
	}

	cmd := &cobra.Command{
		Use:     "branches",
		Short:   "show all review branches",
		Aliases: []string{"b"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if err := parentPersistentPreRunE(cmd, args); err != nil {
				return err
			}

			return c.config.Add(git.FeatureBranch(featureBranch)).Validate(cmd.Context())
		},
		RunE: c.run,
	}
	cmd.Flags().StringVarP(&featureBranch, "feature", "f", "", "feature branch")

	return cmd
}

func (c *branchesCommand) run(cmd *cobra.Command, args []string) error {
	featureBranch, err := c.config.FeatureBranch()
	if err != nil {
		return err
	}

	reviewBranches, err := git.AllReviewBranches(cmd.Context(), featureBranch)
	if err != nil {
		return err
	}

	for _, branch := range reviewBranches {
		fmt.Printf("%s\n", branch.BranchName())
	}

	return nil
}
