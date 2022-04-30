package main

import (
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

type rebaseCommand struct {
	branches *branches
}

func makeRebaseCommand(branches *branches) *cobra.Command {
	c := rebaseCommand{
		branches: branches,
	}

	return &cobra.Command{
		Use:     "rebase",
		Short:   "rebase feature branch",
		Aliases: []string{"r"},
		// PersistentPreRunE не нужен, см. main
		RunE: c.run,
	}
}

func (c *rebaseCommand) run(cmd *cobra.Command, args []string) error {
	return git.Rebase(cmd.Context(), c.branches.baseBranch, c.branches.featureBranch)
}
