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
	// PersistentPreRunE не нужен, см. main
	RunE: listFeatureCommits,
}

const (
	Yellow = "\033[33m"
	Green  = "\033[32m"
	Reset  = "\033[0m"
	Red    = "\033[31m"

	MarkOldCommit     = Red + "--" + Reset
	MarkNewCommit     = Yellow + "++" + Reset
	MarkOkCommit      = Green + "ok" + Reset
	MarkMatchedCommit = Yellow + "**" + Reset
)

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

		switch {
		case record.IsNewCommit():
			fmt.Printf("%d) %s %s %s\n", i+1, MarkNewCommit, record.FeatureSHA, record.FeatureMsg.Subject)
		case record.IsOldCommit():
			fmt.Printf("%d) %s %s [%s] %s\n", i+1, MarkOldCommit, record.ReviewSHA, record.ReviewBranch, record.ReviewMsg.Subject)
		case record.FeatureSHA != record.ReviewSHA:
			fmt.Printf("%d) %s %s [%s] %s\n", i+1, MarkMatchedCommit, record.FeatureSHA, record.ReviewBranch, record.FeatureMsg.Subject)
		default:
			fmt.Printf("%d) %s %s [%s] %s\n", i+1, MarkOkCommit, record.FeatureSHA, record.ReviewBranch, record.FeatureMsg.Subject)
		}
	}

	return nil
}
