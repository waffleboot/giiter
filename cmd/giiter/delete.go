package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/app"
	"github.com/waffleboot/giiter/internal/git"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "delete review branches",
	Aliases: []string{"d"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if app.Config.FeatureBranch == "" {
			return errors.New("feature branch is required")
		}
		return nil
	},
	RunE: deleteReviewBranches,
}

func deleteReviewBranches(cmd *cobra.Command, args []string) error {
	branches, err := git.AllBranches(cmd.Context())
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("review/%s/", app.Config.FeatureBranch)

	for _, branch := range branches {
		if !strings.HasPrefix(branch.BranchName, prefix) {
			continue
		}
		if err := git.DeleteBranch(cmd.Context(), branch.BranchName); err != nil {
			return err
		}
	}

	return nil
}