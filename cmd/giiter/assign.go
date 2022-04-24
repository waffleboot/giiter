package main

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

var assignCmd = &cobra.Command{
	Use:     "assign",
	Short:   "reassign commit to review branch",
	Aliases: []string{"a"},
	// PersistentPreRunE не нужен, см. main
	RunE: assign,
}

func assign(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("need new commit and old review branch position numbers")
	}

	shaPos, branchPos := args[0], args[1]

	shaIndex, err := strconv.Atoi(shaPos)
	if err != nil {
		return err
	}

	branchIndex, err := strconv.Atoi(branchPos)
	if err != nil {
		return err
	}

	records, err := git.State(
		cmd.Context(),
		_baseBranch,
		_featureBranch)
	if err != nil {
		return err
	}

	switch {
	case shaIndex < 0:
		return errors.New("commit position is negative")
	case shaIndex > len(records):
		return errors.New("commit position is greater then count of records")
	case branchIndex < 0:
		return errors.New("branch position is negative")
	case branchIndex > len(records):
		return errors.New("branch position is greater then count of records")
	case shaIndex == branchIndex:
		return errors.New("you point the same record")
	}

	featureSHA := records[shaIndex-1].FeatureSHA
	branchName := records[branchIndex-1].ReviewBranch.BranchName

	reviewIndex := -1
	commitIndex := -1

	for i := range records {
		if records[i].ReviewBranch.BranchName == branchName {
			reviewIndex = i
		}

		if records[i].FeatureSHA == featureSHA {
			commitIndex = i
		}
	}

	// TODO зачем это сделано?

	if reviewIndex < 0 || commitIndex < 0 || reviewIndex == commitIndex {
		return nil
	}

	if err := git.SwitchBranch(cmd.Context(), branchName, featureSHA); err != nil {
		return err
	}

	return listFeatureCommits(cmd, args)
}
