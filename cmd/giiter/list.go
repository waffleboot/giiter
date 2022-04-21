package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "show feature commits",
	Aliases: []string{"l"},
	RunE:    listFeatureCommits,
}

func listFeatureCommits(cmd *cobra.Command, args []string) error {
	records, err := git.State(
		cmd.Context(),
		baseBranch,
		featureBranch)
	if err != nil {
		return err
	}

	for i := range records {
		record := records[i]
		if record.IsNewCommit() {
			fmt.Printf("%d) !! %s %s\n", i+1, record.FeatureSHA, record.FeatureMsg.Subject)
		} else if record.IsOldCommit() {
			fmt.Printf("%d) ?? %s [%s] %s\n", i+1, record.ReviewSHA, record.ReviewBranch, record.ReviewMsg.Subject)
		} else if record.FeatureSHA != record.ReviewSHA {
			fmt.Printf("%d) ** %s [%s] %s\n", i+1, record.FeatureSHA, record.ReviewBranch, record.FeatureMsg.Subject)
		} else {
			fmt.Printf("%d) ok %s [%s] %s\n", i+1, record.FeatureSHA, record.ReviewBranch, record.FeatureMsg.Subject)
		}
	}

	return nil
}
