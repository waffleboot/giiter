package main

import (
	"github.com/spf13/cobra"

	"github.com/waffleboot/giiter/internal/git"
)

func addCommonFlags(cmd *cobra.Command, config *git.Config) {
	var baseBranch, featureBranch string

	cmd.Flags().StringVarP(&baseBranch, "base", "b", "", "base branch")
	cmd.Flags().StringVarP(&featureBranch, "feature", "f", "", "feature branch")
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		return config.Add(
			git.BaseBranch(baseBranch),
			git.FeatureBranch(featureBranch)).
			Validate(cmd.Context())
	}
}
