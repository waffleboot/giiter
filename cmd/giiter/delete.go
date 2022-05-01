package main

import (
	"github.com/spf13/cobra"

	"github.com/waffleboot/giiter/internal/git"
)

type deleteCommand struct {
	config *git.Config
}

func makeDeleteCommand(config *git.Config) *cobra.Command {
	var featureBranch string

	c := deleteCommand{
		config: config,
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete review branches",
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

func (c *deleteCommand) run(cmd *cobra.Command, args []string) error {
	featureBranch, err := c.config.FeatureBranch()
	if err != nil {
		return err
	}

	reviewBranches, err := git.AllReviewBranches(cmd.Context(), featureBranch)
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
