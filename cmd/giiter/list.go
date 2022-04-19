package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/app"
	"github.com/waffleboot/giiter/internal/git"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "show feature commits",
	Aliases: []string{"l"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if app.Config.BaseBranch == "" {
			return errors.New("base branch is required")
		}
		if app.Config.FeatureBranch == "" {
			return errors.New("feature branch is required")
		}
		return nil
	},
	RunE: listFeatureCommits,
}

func listFeatureCommits(cmd *cobra.Command, args []string) error {
	records, err := git.Refresh(cmd.Context(), app.Config.BaseBranch)
	if err != nil {
		return err
	}

	for i := range records {
		record := records[i]
		if record.IsNewCommit() {
			fmt.Printf("%d) + %s %s\n", i+1, record.FeatureSHA, record.FeatureMsg.Subject)
		} else if record.IsOldCommit() {
			fmt.Printf("%d) - %s [%s] %s\n", i+1, record.ReviewSHA, record.ReviewBranch, record.ReviewMsg.Subject)
		} else {
			fmt.Printf("%d) . %s [%s] %s\n", i+1, record.FeatureSHA, record.ReviewBranch, record.FeatureMsg.Subject)
		}
	}

	return nil
}
