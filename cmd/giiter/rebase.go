package main

import (
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

var rebaseCmd = &cobra.Command{
	Use:     "rebase",
	Short:   "rebase feature branch",
	Aliases: []string{"r"},
	RunE:    rebaseFeatureBranch,
}

func rebaseFeatureBranch(cmd *cobra.Command, args []string) error {
	return git.Rebase(cmd.Context(), baseBranch, featureBranch)
}
