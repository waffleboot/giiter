package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/app"
	"github.com/waffleboot/giiter/internal/git"
)

var makeCmd = &cobra.Command{
	Use:     "make",
	Short:   "make review branches",
	Aliases: []string{"m"},
	// PersistentPreRunE не нужен, см. main
	RunE: makeReviewBranches,
}

func makeReviewBranches(cmd *cobra.Command, args []string) error {
	records, err := git.Refresh(
		cmd.Context(),
		_baseBranch,
		_featureBranch)
	if err != nil {
		return err
	}

	prevBranch := _baseBranch

	for i := range records {
		if records[i].HasReview() {
			prevBranch, err = records[i].AnyReviewBranch()
			if err != nil {
				return fmt.Errorf("error on record %d: %s", i+1, err)
			}

			continue
		}

		newBranch := fmt.Sprintf("review/%s/%d", _featureBranch, records[i].NewID)

		title := "Draft: "
		if app.Config.MergeRequestPrefix != "" {
			title += app.Config.MergeRequestPrefix + ": "
		}

		title += records[i].FeatureMsg.Subject

		if err := git.CreateBranch(
			cmd.Context(),
			git.Branch{
				CommitSHA:  records[i].FeatureSHA,
				BranchName: newBranch,
			}); err != nil {
			return err
		}

		if err := git.CreateMergeRequest(
			cmd.Context(),
			git.MergeRequest{
				Title:        title,
				SourceBranch: newBranch,
				TargetBranch: prevBranch,
				Description:  records[i].FeatureMsg.Description,
			}); err != nil {
			return err
		}

		prevBranch = newBranch
	}

	return listFeatureCommits(cmd, args)
}
