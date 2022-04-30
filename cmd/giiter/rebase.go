package main

import (
	"github.com/spf13/cobra"

	"github.com/waffleboot/giiter/internal/git"
)

type rebaseCommand struct {
	config *git.Config
}

func makeRebaseCommand(config *git.Config) *cobra.Command {
	c := rebaseCommand{
		config: config,
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
	baseBranch, featureBranch, err := c.config.Branches()
	if err != nil {
		return err
	}

	return git.Rebase(cmd.Context(), baseBranch, featureBranch)
}
