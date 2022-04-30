package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/app"
	"github.com/waffleboot/giiter/internal/git"
)

type makeCommand struct {
	branches *branches
}

func makeMakeCommand(branches *branches) *cobra.Command {
	c := makeCommand{
		branches: branches,
	}
	cmd := &cobra.Command{
		Use:     "make",
		Short:   "make review branches",
		Aliases: []string{"m"},
		// PersistentPreRunE не нужен, см. main
		RunE: c.run,
	}
	cmd.Flags().StringVarP(&app.Config.MergeRequestPrefix, "prefix", "t", "", "title prefix for merge request")

	return cmd
}

func (c *makeCommand) run(cmd *cobra.Command, args []string) error {
	records, err := git.Refresh(
		cmd.Context(),
		c.branches.baseBranch,
		c.branches.featureBranch)
	if err != nil {
		return err
	}

	prevBranch := c.branches.baseBranch

	for i := range records {
		if records[i].HasReview() {
			prevBranch, err = records[i].AnyReviewBranch()
			if err != nil {
				return fmt.Errorf("error on record %d: %s", i+1, err)
			}

			continue
		}

		newBranch := fmt.Sprintf(git.Prefix+"%s/%d", c.branches.featureBranch, records[i].NewID)

		title := "Draft: "
		if app.Config.MergeRequestPrefix != "" {
			title += app.Config.MergeRequestPrefix + ": "
		}

		title += records[i].CommitMessage().Subject

		if err := git.CreateBranch(
			cmd.Context(),
			git.Branch{
				CommitSHA:  records[i].CommitSHA(),
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
				Description:  records[i].CommitMessage().Description,
			}); err != nil {
			return err
		}

		prevBranch = newBranch
	}

	return listFeatureCommits(cmd.Context(), c.branches)
}
