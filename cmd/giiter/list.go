package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/waffleboot/giiter/internal/git"
)

type listCommand struct {
	config *git.Config
}

func makeListCommand(config *git.Config) *cobra.Command {
	c := listCommand{
		config: config,
	}

	return &cobra.Command{
		Use:     "list",
		Short:   "show feature commits",
		Aliases: []string{"l"},
		// PersistentPreRunE не нужен, см. main
		RunE: c.run,
	}
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

func (c *listCommand) run(cmd *cobra.Command, args []string) error {
	return listFeatureCommits(cmd.Context(), c.config)
}

func listFeatureCommits(ctx context.Context, c *git.Config) error {
	baseBranch, featureBranch, err := c.Branches()
	if err != nil {
		return err
	}

	records, err := git.State(ctx, baseBranch, featureBranch)
	if err != nil {
		return err
	}

	for i := range records {
		record := records[i]

		commitSHA := record.CommitSHA()
		commitMsg := Yellow + record.CommitMessage().Subject + Reset
		reviewBranches := strings.Join(record.ReviewBranchNamesForUI(), ",")

		switch {
		case record.IsNewCommit():
			fmt.Printf("%d) %s %s %s\n", i+1,
				MarkNewCommit,
				commitSHA,
				commitMsg)
		case record.IsOldCommit():
			fmt.Printf("%d) %s %s [%s] %s\n", i+1,
				MarkOldCommit,
				commitSHA,
				reviewBranches,
				commitMsg)
		case record.MatchedCommit():
			fmt.Printf("%d) %s %s [%s] %s\n", i+1,
				MarkOkCommit,
				commitSHA,
				reviewBranches,
				commitMsg)
		default:
			fmt.Printf("%d) %s %s [%s] %s\n", i+1,
				MarkSwitchCommit,
				commitSHA,
				reviewBranches,
				commitMsg)
		}
	}

	return nil
}
