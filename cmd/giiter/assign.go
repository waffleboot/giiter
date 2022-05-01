package main

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/waffleboot/giiter/internal/git"
)

type assignCommand struct {
	config *git.Config
}

func makeAssignCommand(config *git.Config) *cobra.Command {
	c := assignCommand{
		config: config,
	}

	return &cobra.Command{
		Use:     "assign",
		Short:   "reassign commit to review branch",
		Aliases: []string{"a"},
		// PersistentPreRunE не нужен, см. main
		RunE: c.run,
	}
}

func (c *assignCommand) run(cmd *cobra.Command, args []string) error {
	baseBranch, featureBranch, err := c.config.Branches()
	if err != nil {
		return err
	}

	if len(args) < 2 {
		return errors.New("need new commit and old review branch position numbers")
	}

	commitPos, branchPos := args[0], args[1]

	commitIndex, err := strconv.Atoi(commitPos)
	if err != nil {
		return err
	}

	branchIndex, err := strconv.Atoi(branchPos)
	if err != nil {
		return err
	}

	records, err := git.State(cmd.Context(), baseBranch, featureBranch)
	if err != nil {
		return err
	}

	commitIndex--
	branchIndex--

	switch {
	case commitIndex < 0:
		return errors.New("commit position is negative")
	case commitIndex > len(records):
		return errors.New("commit position is greater then count of records")
	case branchIndex < 0:
		return errors.New("branch position is negative")
	case branchIndex > len(records):
		return errors.New("branch position is greater then count of records")
	case commitIndex == branchIndex:
		return errors.New("you point the same record")
	case records[commitIndex].HasReview():
		return fmt.Errorf("could not reassign commit %s with review", commitPos)
	case !records[branchIndex].HasReview():
		return fmt.Errorf("could not reassign commit %s without review", branchPos)
	}

	commit := records[commitIndex].CommitSHA()

	branchName, err := records[branchIndex].AnyReviewBranch()
	if err != nil {
		return err
	}

	if err := git.SwitchBranch(cmd.Context(), branchName, commit); err != nil {
		return err
	}

	return listFeatureCommits(cmd.Context(), c.config)
}
