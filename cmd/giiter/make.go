package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/app"
	"github.com/waffleboot/giiter/internal/git"
)

var makeCmd = &cobra.Command{
	Use:     "make",
	Short:   "make review branches",
	Aliases: []string{"m"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if app.Config.BaseBranch == "" {
			return errors.New("base branch is required")
		}
		if app.Config.FeatureBranch == "" {
			return errors.New("feature branch is required")
		}
		return nil
	},
	RunE: makeReviewBranches,
}

func makeReviewBranches(cmd *cobra.Command, args []string) error {
	records, err := g.Refresh(cmd.Context())
	if err != nil {
		return err
	}

	for i := range records {

		if records[i].ReviewSHA != "" {
			continue
		}

		baseBranch := app.Config.BaseBranch
		if i > 0 {
			baseBranch = records[i-1].ReviewBranch
		}

		newBranch := fmt.Sprintf("review/%s/%d", app.Config.FeatureBranch, records[i].ID)

		title := "Draft: "
		if app.Config.Prefix != "" {
			title += app.Config.Prefix + ": "
		}
		title += records[i].FeatureMsg.Subject

		if err := g.CreateBranch(
			cmd.Context(),
			git.CreateBranchRequest{
				SHA:         records[i].FeatureSHA,
				Branch:      newBranch,
				Target:      baseBranch,
				Title:       title,
				Description: records[i].FeatureMsg.Description,
			}); err != nil {
			return err
		}

		records[i].ReviewBranch = newBranch

	}

	return listFeatureCommits(cmd, args)
}
