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
		baseBranch,
		featureBranch)
	if err != nil {
		return err
	}

	for i := range records {
		if records[i].ReviewSHA != "" {
			continue
		}

		prevBranch := baseBranch
		if i > 0 {
			prevBranch = records[i-1].ReviewBranch
		}

		newBranch := fmt.Sprintf("review/%s/%d", featureBranch, records[i].ID)

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

		records[i].ReviewBranch = newBranch
	}

	return listFeatureCommits(cmd, args)
}
