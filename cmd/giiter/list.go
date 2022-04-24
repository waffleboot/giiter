package main

import (
	"fmt"
	"strings"

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

	MarkSwitchCommit = Yellow + "**" + Reset
	MarkNewCommit    = Yellow + "++" + Reset
	MarkOldCommit    = Red + "--" + Reset
	MarkOkCommit     = Green + "ok" + Reset
)

func listFeatureCommits(cmd *cobra.Command, args []string) error {
	records, err := git.State(
		cmd.Context(),
		_baseBranch,
		_featureBranch)
	if err != nil {
		return err
	}

	for i := range records {
		record := records[i]

		switch {
		case record.IsNewCommit():
			fmt.Printf("%d) %s %s %s\n", i+1,
				MarkNewCommit,
				record.CommitSHA(),
				record.CommitMessage().Subject)
		case record.IsOldCommit():
			fmt.Printf("%d) %s %s [%s] %s\n", i+1,
				MarkOldCommit,
				record.CommitSHA(),
				strings.Join(record.ReviewBranchNames(), ","),
				record.CommitMessage().Subject)
		case record.MatchedCommit():
			fmt.Printf("%d) %s %s [%s] %s\n", i+1,
				MarkOkCommit,
				record.CommitSHA(),
				strings.Join(record.ReviewBranchNames(), ","),
				record.CommitMessage().Subject)
		default:
			fmt.Printf("%d) %s %s [%s] %s\n", i+1,
				MarkSwitchCommit,
				record.CommitSHA(),
				strings.Join(record.ReviewBranchNames(), ","),
				record.CommitMessage().Subject)
		}
	}

	return nil
}
