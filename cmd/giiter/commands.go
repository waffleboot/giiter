package main

import (
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

type branches struct {
	baseBranch    string
	featureBranch string
}

func (c *branches) addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&c.baseBranch, "base", "b", "", "base branch")
	cmd.Flags().StringVarP(&c.featureBranch, "feature", "f", "", "feature branch")
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		c.baseBranch, c.featureBranch, err = git.FindBaseAndFeatureBranches(cmd.Context(), c.baseBranch, c.featureBranch)

		return
	}
}
